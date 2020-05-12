import * as React from "react";
import { Assignment, Course, User } from "../../../proto/ag_pb";
import { DynamicTable, Row, Search, StudentLab } from "../../components";
import { IStudentLabsForCourse, IStudentLab, ISubmission } from "../../models";
import { ICellElement } from "../data/DynamicTable";
import { generateCellClass, sortByScore } from "./labHelper";
import { getSlipDays, groupRepoLink, searchForLabs } from "../../componentHelper";

interface IResultsProps {
    course: Course;
    courseURL: string;
    groups: IStudentLabsForCourse[];
    labs: Assignment[];
    getReviewers: (submissionID: number) => Promise<string[]>;
    onApproveClick: (submissionID: number, approved: boolean) => Promise<boolean>;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<ISubmission | null>;
}

interface IResultsState {
    submissionLink?: IStudentLab;
    groups: IStudentLabsForCourse[];
}

export class GroupResults extends React.Component<IResultsProps, IResultsState> {

    constructor(props: IResultsProps) {
        super(props);

        const currentGroup = this.props.groups.length > 0 ? this.props.groups[0] : null;
        const allAssignments = currentGroup ? currentGroup.course.getAssignmentsList() : null;
        if (currentGroup && allAssignments && allAssignments.length > 0) {
            this.state = {
                // Only using the first group to fetch assignments.
                submissionLink: currentGroup.labs[0],
                groups: sortByScore(this.props.groups, this.props.labs, true),
            };
        } else {
            this.state = {
                submissionLink: undefined,
                groups: sortByScore(this.props.groups, this.props.labs, true),
            };
        }
    }

    public render() {
        let groupLab: JSX.Element | null = null;
        const currentGroups = this.props.groups.length > 0 ? this.props.groups : null;
        if (currentGroups
            && this.state.submissionLink
            && this.state.submissionLink.assignment.getIsgrouplab()) {
            groupLab = <StudentLab
                studentSubmission={this.state.submissionLink}
                student={new User()}
                courseURL={this.props.courseURL}
                teacherPageView={false}
                courseCreatorView={false}
                slipdays={this.props.course.getSlipdays()}
                showApprove={true}
                getReviewers={this.props.getReviewers}
                onRebuildClick={
                    async () => {
                        if (this.state.submissionLink && this.state.submissionLink.submission) {
                            const ans = await this.props.onRebuildClick(this.state.submissionLink.assignment.getId(), this.state.submissionLink.submission.id);
                            if (ans) {
                                this.state.submissionLink.submission = ans;
                                return true;
                            }
                        }
                        return false;
                    }
                }
                onApproveClick={(approve: boolean) => {
                    if (this.state.submissionLink && this.state.submissionLink.submission) {
                        const ans = this.props.onApproveClick(this.state.submissionLink.submission.id, approve);
                        if (ans) {
                            this.state.submissionLink.submission.approved = approve;
                        }
                    }
                }}
            />;
        }

        return (
            <div>
                <h1>Result: {this.props.course.getName()}</h1>
                <Row>
                    <div key="resulthead" className="col-lg6 col-md-6 col-sm-12">
                        <Search className="input-group"
                            placeholder="Search for groups"
                            onChange={(query) => this.handleSearch(query)}
                        />
                        <DynamicTable header={this.getResultHeader()}
                            data={this.state.groups}
                            selector={(item: IStudentLabsForCourse) => this.getGroupResultSelector(item)}
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

    private getGroupResultSelector(group: IStudentLabsForCourse): (string | JSX.Element | ICellElement)[] {
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

    private handleOnclick(item: IStudentLab): void {
        this.setState({
            submissionLink: item,
        });
    }

    private handleSearch(query: string): void {
        this.setState({
            groups: searchForLabs(this.props.groups, query),
        });
    }
}
