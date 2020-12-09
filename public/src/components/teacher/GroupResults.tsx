import * as React from "react";
import { Assignment, Course, Submission, User, Comment } from '../../../proto/ag_pb';
import { DynamicTable, Row, Search, StudentLab } from "../../components";
import { IAllSubmissionsForEnrollment, ISubmissionLink, ISubmission } from "../../models";
import { ICellElement } from "../data/DynamicTable";
import { generateCellClass, sortByScore } from "./labHelper";
import { groupRepoLink, searchForLabs } from "../../componentHelper";

interface IResultsProps {
    course: Course;
    courseURL: string;
    allGroupSubmissions: IAllSubmissionsForEnrollment[];
    assignments: Assignment[];
    updateSubmissionStatus: (submission: ISubmission) => Promise<boolean>;
    updateComment: (comment: Comment) => Promise<boolean>;
    deleteComment: (commentID: number) => void;
    rebuildSubmission: (assignmentID: number, submissionID: number) => Promise<ISubmission | null>;
}

interface IResultsState {
    commenting: boolean;
    selectedSubmission?: ISubmissionLink;
    groups: IAllSubmissionsForEnrollment[];
}

export class GroupResults extends React.Component<IResultsProps, IResultsState> {

    constructor(props: IResultsProps) {
        super(props);

        const currentGroup = this.props.allGroupSubmissions.length > 0 ? this.props.allGroupSubmissions[0] : null;
        const allAssignments = currentGroup ? currentGroup.course.getAssignmentsList() : null;
        if (currentGroup && allAssignments && allAssignments.length > 0) {
            this.state = {
                commenting: false,
                // Only using the first group to fetch assignments.
                selectedSubmission: currentGroup.labs[0],
                groups: sortByScore(this.props.allGroupSubmissions, this.props.assignments, true),
            };
        } else {
            this.state = {
                commenting: false,
                selectedSubmission: undefined,
                groups: sortByScore(this.props.allGroupSubmissions, this.props.assignments, true),
            };
        }
    }

    public render() {
        let groupLab: JSX.Element | null = null;
        const currentGroups = this.props.allGroupSubmissions.length > 0 ? this.props.allGroupSubmissions : null;
        if (currentGroups
            && this.state.selectedSubmission
            && this.state.selectedSubmission.assignment.getIsgrouplab()) {
            groupLab = <StudentLab
                submissionLink={this.state.selectedSubmission}
                student={new User()}
                courseURL={this.props.courseURL}
                teacherPageView={true}
                slipdays={this.props.course.getSlipdays()}
                commenting={this.state.commenting}
                rebuildSubmission={ () => this.rebuildSubmission()}
                updateSubmissionStatus={(status: Submission.Status) => this.updateSubmissionStatus(status)}
                updateComment={(comment: Comment) => this.setSubmissionComment(comment)}
                deleteComment={(commentID: number) => this.props.deleteComment(commentID)}
                toggleCommenting={(on: boolean) => this.toggleCommenting(on)}
            />;
        }

        return (
            <div
            onKeyDown={(e) => {
                if (!this.state.commenting) {
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
        headers = headers.concat(this.props.assignments.filter((e) => e.getIsgrouplab()).map((e) => e.getName()));
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
                        style={{ whiteSpace: 'nowrap' }}
                        onClick={() => this.handleOnclick(e)}
                        href="#">
                        {e.submission ? (e.submission.score + " %") : "N/A"}</a>,
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
            const ans = await this.props.updateSubmissionStatus(selected);
            if (!ans) {
                selected.status = previousStatus;
            }
            this.setState({
                selectedSubmission: current,
            });
        }
    }

    private async setSubmissionComment(comment: Comment) {
        const selected = this.state.selectedSubmission?.submission;
        if (selected) {
            const current = this.state.selectedSubmission;
            const ans = await this.props.updateComment(comment);
            if (ans) {
                this.setState({
                    selectedSubmission: current,
                });
            }
        }
    }

    private async rebuildSubmission(): Promise<boolean> {
        const currentSubmission = this.state.selectedSubmission;
        if (currentSubmission && currentSubmission.submission) {
            const ans = await this.props.rebuildSubmission(currentSubmission.assignment.getId(), currentSubmission.submission.id);
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
            groups: sortByScore(searchForLabs(this.props.allGroupSubmissions, query), this.props.assignments, true),
        });
    }

    private toggleCommenting(toggleOn: boolean) {
        this.setState({
            commenting: toggleOn,
        })
    }
}
