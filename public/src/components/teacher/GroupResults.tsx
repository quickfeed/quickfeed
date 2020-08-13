import * as React from "react";
import { Assignment, Course, Submission, User } from "../../../proto/ag_pb";
import { DynamicTable, Row, Search, StudentLab } from "../../components";
import { IAllSubmissionsForEnrollment, ISubmissionLink, ISubmission } from "../../models";
import { ICellElement } from "../data/DynamicTable";
import { generateCellClass, sortByScore } from "./labHelper";
import { groupRepoLink, searchForLabs } from "../../componentHelper";

interface IResultsProps {
    course: Course;
    courseURL: string;
    groups: IAllSubmissionsForEnrollment[];
    labs: Assignment[];
    onSubmissionStatusUpdate: (submission: ISubmission) => Promise<boolean>;
    onSubmissionRebuild: (assignmentID: number, submissionID: number) => Promise<ISubmission | null>;
}

interface IResultsState {
    selectedSubmission?: ISubmissionLink;
    groups: IAllSubmissionsForEnrollment[];
}

export class GroupResults extends React.Component<IResultsProps, IResultsState> {

    constructor(props: IResultsProps) {
        super(props);

        const currentGroup = this.props.groups.length > 0 ? this.props.groups[0] : null;
        const allAssignments = currentGroup ? currentGroup.course.getAssignmentsList() : null;
        if (currentGroup && allAssignments && allAssignments.length > 0) {
            this.state = {
                // Only using the first group to fetch assignments.
                selectedSubmission: currentGroup.labs[0],
                groups: sortByScore(this.props.groups, this.props.labs, true),
            };
        } else {
            this.state = {
                selectedSubmission: undefined,
                groups: sortByScore(this.props.groups, this.props.labs, true),
            };
        }
    }

    public render() {
        let groupLab: JSX.Element | null = null;
        const currentGroups = this.props.groups.length > 0 ? this.props.groups : null;
        if (currentGroups
            && this.state.selectedSubmission
            && this.state.selectedSubmission.assignment.getIsgrouplab()) {
            groupLab = <StudentLab
                studentSubmission={this.state.selectedSubmission}
                student={new User()}
                courseURL={this.props.courseURL}
                teacherPageView={true}
                slipdays={this.props.course.getSlipdays()}
                onSubmissionRebuild={ () => this.rebuildSubmission()}
                onSubmissionStatusUpdate={(status: Submission.Status) => this.updateSubmissionStatus(status)}
            />;
        }

        return (
            <div
            onKeyDown={(e) => {
                switch (e.key) {
                    case "a": {
                        this.updateSubmissionStatus(Submission.Status.APPROVED);
                        break;
                    }
                    case "r": {
                        this.updateSubmissionStatus(Submission.Status.REVISION);
                        break;
                    }
                    case "f": {
                        this.updateSubmissionStatus(Submission.Status.REJECTED)
                        break;
                    }
                    case "b": {
                        this.rebuildSubmission();
                        break;
                    }
                }
            }}
            >
                <h1>Result: {this.props.course.getName()}</h1>
                <Row>
                    <div key="resulthead" className="col-lg6 col-md-6 col-sm-12">
                        <Search className="input-group"
                            placeholder="Search for groups"
                            onChange={(query) => this.handleSearch(query)}
                        />
                        <DynamicTable header={this.getResultHeader()}
                            data={this.state.groups}
                            selector={(item: IAllSubmissionsForEnrollment) => this.getGroupResultSelector(item)}
                        />
                    </div>
                    <div key="resultbody" className="col-lg-6 col-md-6 col-sm-12">
                        {groupLab}
                    </div>
                </Row>
            </div>
        );
    }

    private getResultHeader(): string[] {
        let headers: string[] = ["Name"];
        headers = headers.concat(this.props.labs.filter((e) => e.getIsgrouplab()).map((e) => e.getName()));
        return headers;
    }

    private getGroupResultSelector(group: IAllSubmissionsForEnrollment): (string | JSX.Element | ICellElement)[] {
        const grp = group.enrollment.getGroup();
        const name = grp ? groupRepoLink(grp.getName(), this.props.courseURL) : "";
        let selector: (string | JSX.Element | ICellElement)[] = [name];
        selector = selector.concat(group.labs.filter((e, i) => e.assignment.getIsgrouplab()).map(
            (e, i) => {
                let cellCss: string = "";
                if (e.submission) {
                    cellCss = generateCellClass(e);
                }
                const iCell: ICellElement = {
                    value: <a className={cellCss + " lab-cell-link"}
                        onClick={() => this.handleOnclick(e)}
                        href="#">
                        {e.submission ? (e.submission.score + "%") : "N/A"}</a>,
                    className: cellCss,
                };
                return iCell;
            }));
        return selector;
    }

    private async updateSubmissionStatus(status: Submission.Status) {
        const current = this.state.selectedSubmission;
        const selected = current?.submission;
        if (selected) {
            const previousStatus = selected.status;
            selected.status = status;
            const ans = await this.props.onSubmissionStatusUpdate(selected);
            if (!ans) {
                selected.status = previousStatus;
            }
            this.setState({
                selectedSubmission: current,
            });
        }
    }

    private async rebuildSubmission(): Promise<boolean> {
        const currentSubmission = this.state.selectedSubmission;
        if (currentSubmission && currentSubmission.submission) {
            const ans = await this.props.onSubmissionRebuild(currentSubmission.assignment.getId(), currentSubmission.submission.id);
            if (ans) {
                currentSubmission.submission = ans;
                this.setState({
                    selectedSubmission: currentSubmission,
                });
                return true;
            }
        }
        return false;
    }

    private async handleOnclick(item: ISubmissionLink) {
        this.setState({
            selectedSubmission: item,
        });
    }

    private handleSearch(query: string): void {
        this.setState({
            groups: searchForLabs(this.props.groups, query),
        });
    }
}
