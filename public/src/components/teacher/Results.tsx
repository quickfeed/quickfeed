import * as React from "react";
import { Assignment, Course, User, Submission } from "../../../proto/ag_pb";
import { DynamicTable, Row, Search, StudentLab } from "../../components";
import { IAllSubmissionsForEnrollment, ISubmissionLink, ISubmission } from "../../models";
import { ICellElement } from "../data/DynamicTable";
import { generateCellClass, sortByScore } from "./labHelper";
import { searchForLabs, userRepoLink, getSlipDays, legalIndex } from "../../componentHelper";

interface IResultsProps {
    course: Course;
    courseURL: string;
    allCourseSubmissions: IAllSubmissionsForEnrollment[];
    assignments: Assignment[];
    courseCreatorView: boolean;
    onSubmissionStatusUpdate: (submission: ISubmission) => Promise<boolean>;
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
                // Only using the first student to fetch assignments.
                selectedSubmission: currentStudent.labs[0],
                allSubmissions: sortByScore(this.props.allCourseSubmissions, this.props.assignments, false),
                ignoreShortcuts: false,
            };
        } else {
            this.state = {
                selectedSubmission: undefined,
                allSubmissions: sortByScore(this.props.allCourseSubmissions, this.props.assignments, false),
                ignoreShortcuts: false,
            };
        }
    }

    public render() {
        let studentLab: JSX.Element | null = null;
        const currentStudents = this.props.allCourseSubmissions.length > 0 ? this.props.allCourseSubmissions : null;
        if (currentStudents
            && this.state.selectedSubmission && this.state.selectedStudent
            && !this.state.selectedSubmission.assignment.getIsgrouplab()
        ) {
            studentLab = <StudentLab
                studentSubmission={this.state.selectedSubmission}
                courseURL={this.props.courseURL}
                student={this.state.selectedStudent.enrollment.getUser() ?? new User()}
                teacherPageView={true}
                slipdays={this.state.selectedSubmission.submission ? getSlipDays(this.props.allCourseSubmissions, this.state.selectedSubmission.submission, false) : 0}
                onSubmissionRebuild={() => this.rebuildSubmission()}
                onSubmissionStatusUpdate={(status: Submission.Status) => this.updateSubmissionStatus(status)}
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
                <h1>Result: {this.props.course.getName()}</h1>
                <Row>
                    <div key="resultshead" className="col-lg6 col-md-6 col-sm-12">
                        <Search className="input-group"
                            placeholder="Search for students"
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
                            selector={(item: IAllSubmissionsForEnrollment) => this.getResultSelector(item)}
                        />
                    </div>
                    <div key="resultsbody" className="col-lg-6 col-md-6 col-sm-12">
                        {studentLab}
                    </div>
                </Row>
            </div>
        );
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

    private getResultHeader(): string[] {
        let headers: string[] = ["Name"];
        headers = headers.concat(this.props.assignments.filter((e) => !e.getIsgrouplab()).map((e) => e.getName()));
        return headers;
    }

    private getResultSelector(student: IAllSubmissionsForEnrollment): (string | JSX.Element | ICellElement)[] {
        const user = student.enrollment.getUser();
        const displayName = user ? userRepoLink(user.getLogin(), user.getName(), this.props.courseURL) : "";
        let selector: (string | JSX.Element | ICellElement)[] = [displayName];
        selector = selector.concat(student.labs.filter((e, i) => !e.assignment.getIsgrouplab()).map(
            (e) => {
                let cellCss: string = "";
                if (e.submission) {
                    cellCss = generateCellClass(e);
                }
                const iCell: ICellElement = {
                    value: <a className={cellCss + " lab-cell-link"}
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
}
