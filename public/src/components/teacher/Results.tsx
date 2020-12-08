import * as React from "react";
import { Assignment, Comment, Course, User, Submission } from "../../../proto/ag_pb";
import { DynamicTable, Row, Search, StudentLab } from "../../components";
import { IAllSubmissionsForEnrollment, ISubmissionLink, ISubmission } from "../../models";
import { ICellElement } from "../data/DynamicTable";
import { generateCellClass, sortByScore } from "./labHelper";
import { searchForLabs, userRepoLink, getSlipDays, legalIndex, groupRepoLink } from '../../componentHelper';
import { formatDate } from '../../helper';

interface IResultsProps {
    currentUser: number;
    course: Course;
    courseURL: string;
    allCourseSubmissions: IAllSubmissionsForEnrollment[];
    assignments: Assignment[];
    courseCreatorView: boolean;
    updateSubmissionStatus: (submission: ISubmission) => Promise<boolean>;
    updateComment: (comment: Comment) => Promise<boolean>;
    deleteComment: (commentID: number) => void;
    onSubmissionRebuild: (assignmentID: number, submissionID: number) => Promise<ISubmission | null>;
}

interface IResultsState {
    ignoreShortcuts: boolean;
    selectedSubmission?: ISubmissionLink;
    selectedStudent?: IAllSubmissionsForEnrollment;
    allSubmissions: IAllSubmissionsForEnrollment[];
}

export class Results extends React.Component<IResultsProps, IResultsState> {

    constructor(props: IResultsProps) {
        super(props);

        const currentStudent = this.props.allCourseSubmissions.length > 0 ? this.props.allCourseSubmissions[0] : null;
        const courseAssignments = currentStudent ? currentStudent.course.getAssignmentsList() : null;
        if (currentStudent && courseAssignments && courseAssignments.length > 0) {
            this.state = {
                ignoreShortcuts: false,
                // Only using the first student to fetch assignments.
                selectedSubmission: currentStudent.labs[0],
                allSubmissions: sortByScore(this.props.allCourseSubmissions, this.props.assignments, false),
            };
        } else {
            this.state = {
                ignoreShortcuts: false,
                selectedSubmission: undefined,
                allSubmissions: sortByScore(this.props.allCourseSubmissions, this.props.assignments, false),
            };
        }
    }

    public render() {
        let studentLab: JSX.Element | null = null;
        const currentStudents = this.props.allCourseSubmissions.length > 0 ? this.props.allCourseSubmissions : null;
        if (currentStudents
            && this.state.selectedSubmission && this.state.selectedStudent
        ) {
            studentLab = <StudentLab
                studentSubmission={this.state.selectedSubmission}
                courseURL={this.props.courseURL}
                student={this.state.selectedStudent.enrollment.getUser() ?? new User()}
                teacherPageView={true}
                commenting={this.state.ignoreShortcuts}
                slipdays={this.state.selectedSubmission.submission ? getSlipDays(this.props.allCourseSubmissions, this.state.selectedSubmission.submission, false) : 0}
                onSubmissionRebuild={() => this.rebuildSubmission()}
                updateSubmissionStatus={(status: Submission.Status) => this.updateSubmissionStatus(status)}
                updateComment={(comment: Comment) => this.setSubmissionComment(comment)}
                deleteComment={(commentID: number) => this.props.deleteComment(commentID)}
                toggleCommenting={(on: boolean) => this.toggleCommenting(on)}
            />;
        }

        return (
            <div

            onKeyDown={(e) => {
                if (!this.state.ignoreShortcuts) {
                    switch (e.key) {
                        case "ArrowDown": {
                            this.selectNextStudent(false);
                            break;
                        }
                        case "ArrowUp": {
                            this.selectNextStudent(true);
                            break;
                        }
                        case "ArrowRight": {
                            this.selectNextSubmission(false);
                            break;
                        }
                        case "ArrowLeft": {
                            this.selectNextSubmission(true);
                            break;
                        }
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
                <h1>Results: {this.props.course.getName()}</h1>
                <Row>
                    <div key="resultshead" className="col-lg-8 col-md-8 col-sm-12 col-xs-12">
                        <Search className="input-group"
                            placeholder="Search for students and groups by name, email or GitHub login"
                            onChange={(query) => this.handleSearch(query)}
                            onFocus={() => this.setState({
                                ignoreShortcuts: true,
                            })}
                            onBlur={() => this.setState({
                                ignoreShortcuts: false,
                            })}
                        />
                        <DynamicTable header={this.getResultHeader()}
                            data={this.state.allSubmissions}
                            classType={"result-table"}
                            selector={(item: IAllSubmissionsForEnrollment) => this.getResultSelector(item)}
                        />
                    </div>
                    <div key="resultsbody" className="col-lg-4 col-md-4 col-sm-12 col-xs-12">
                        {studentLab}
                    </div>
                </Row>
            </div>
        );
    }

    private async updateSubmissionStatus(status: Submission.Status) {
        const currentSubmissionLink = this.state.selectedSubmission;
        const selectedSubmission = currentSubmissionLink?.submission;
        if (currentSubmissionLink && selectedSubmission) {
            const previousStatus = selectedSubmission.status;
            selectedSubmission.status = status;
            const ans = await this.props.updateSubmissionStatus(selectedSubmission);
            if (ans) {
                selectedSubmission.approvedDate = new Date().toLocaleString();
                // If the submission is for group assignment, every group member will have a copy
                // in their Submission link structures. When the submission has been
                // approved for one student, update all its copies for every group member.

                if (selectedSubmission.groupid > 0) {
                    this.state.allSubmissions.forEach((e) => {
                        if (e.enrollment.getGroup()?.getId() === selectedSubmission.groupid) {
                            e.labs.forEach((l) => {
                                if (l.assignment.getId() === currentSubmissionLink.assignment.getId()) {
                                    const currentSubmissionCopy = l.submission;
                                    if (currentSubmissionCopy) {
                                        currentSubmissionCopy.approvedDate = selectedSubmission.approvedDate;
                                        currentSubmissionCopy.status = selectedSubmission.status;
                                    }
                                }
                            })
                        }
                    })
                }

            } else {
                selectedSubmission.status = previousStatus;
            }
            this.setState({
                selectedSubmission: currentSubmissionLink,
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

    private getResultHeader(): (string | JSX.Element)[] {
        let headers: (string | JSX.Element)[] = ["Name", "Group"];
        headers = headers.concat(this.props.assignments.map((e) => {
            if (e.getIsgrouplab()) {
                return <span style={{ whiteSpace: 'nowrap' }}>{e.getName() + " (g)"}</span>;
            } else {
                return e.getName();
            }
        }));
        return headers;
    }

    private getResultSelector(student: IAllSubmissionsForEnrollment): (string | JSX.Element | ICellElement)[] {
        const user = student.enrollment.getUser();
        const group = student.enrollment.getGroup();
        const displayName = user ? userRepoLink(user.getLogin(), user.getName(), this.props.courseURL) : "";
        const groupName = group ? groupRepoLink(group.getName(), this.props.courseURL) : "";
        let selector: (string | JSX.Element | ICellElement)[] = [displayName, groupName];
        selector = selector.concat(student.labs.map(
            (e) => {
                let cellCss: string = "";
                if (e.submission) {
                    cellCss = generateCellClass(e);
                }
                const iCell: ICellElement = {
                    value: <a className={cellCss + " lab-cell-link"}
                        style={{ whiteSpace: 'nowrap' }}
                        onClick={() => this.handleOnclick(e, student)}
                        href="#">
                        {e.submission ? (e.submission.score + " %") : "N/A"}</a>,
                    className: cellCss,
                };
                return iCell;
            }));
        return selector;
    }

    private async handleOnclick(item: ISubmissionLink, student: IAllSubmissionsForEnrollment) {
        this.setState({
            selectedSubmission: item,
            selectedStudent: student,
            ignoreShortcuts: false,
        });
    }

    private handleSearch(query: string): void {
        this.setState({
            allSubmissions: sortByScore(searchForLabs(this.props.allCourseSubmissions, query), this.props.assignments, false),
        });
    }

    private selectNextStudent(moveUp: boolean) {
        const currentStudent = this.state.selectedStudent;
        if (currentStudent) {
            const indexOfSelectedStudent = this.props.allCourseSubmissions.findIndex((item) => item.enrollment.getId() === currentStudent.enrollment.getId());
            const currentAssignmentID = this.state.selectedSubmission?.assignment.getId() ?? 0;
            if (indexOfSelectedStudent >= 0 && currentAssignmentID > 0) {
                const nextStudentIndex = moveUp ? indexOfSelectedStudent - 1 : indexOfSelectedStudent + 1;
                if (!legalIndex(nextStudentIndex, this.props.allCourseSubmissions.length)) {
                    return;
                }
                const nextStudent = this.props.allCourseSubmissions[nextStudentIndex];
                if (nextStudent) {
                    const nextStudentSubmission = nextStudent.labs.find((item) => item.assignment.getId() === currentAssignmentID);
                    if (nextStudentSubmission) {
                        this.handleOnclick(nextStudentSubmission, nextStudent);
                    }
                }
            }
        }
    }

    private selectNextSubmission(moveLeft: boolean) {
        const currentStudent = this.state.selectedStudent;
        const currentSubmission = this.state.selectedSubmission;
        if (currentStudent && currentSubmission) {
            const currentAssignmentIndex = this.props.assignments.findIndex(item => item.getId() === currentSubmission.assignment.getId());
            if (currentAssignmentIndex >= 0) {
                const nextAssignmentIndex = moveLeft ? currentAssignmentIndex - 1 : currentAssignmentIndex + 1;
                if (!legalIndex(nextAssignmentIndex, this.props.assignments.length)) {
                    return;
                }
                const nextAssignment = this.props.assignments[nextAssignmentIndex];
                if (nextAssignment) {
                    const submissionForNextAssignment = currentStudent.labs.find(item => item.assignment.getId() === nextAssignment.getId());
                    if (submissionForNextAssignment) {
                        this.handleOnclick(submissionForNextAssignment, currentStudent);
                    }
                }
            }
        }
    }

    private toggleCommenting(toggleOn: boolean) {
        this.setState({
            ignoreShortcuts: toggleOn,
        })
    }
}
