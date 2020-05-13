import * as React from "react";
import { Assignment, Course, User, Submission } from '../../../proto/ag_pb';
import { DynamicTable, Row, Search, StudentLab } from "../../components";
import { IStudentLabsForCourse, IStudentLab, ISubmission } from "../../models";
import { ICellElement } from "../data/DynamicTable";
import { generateCellClass, sortByScore } from "./labHelper";
import { searchForLabs, userRepoLink, getSlipDays } from '../../componentHelper';

interface IResultsProp {
    course: Course;
    courseURL: string;
    allCourseSubmissions: IStudentLabsForCourse[];
    assignments: Assignment[];
    courseCreatorView: boolean;
    onApproveClick: (submissionID: number, approve: boolean) => Promise<boolean>;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<ISubmission | null>;
    getReviewers: (submissionID: number) => Promise<string[]>
    setApproved: (submissionID: number, status: Submission.Status) => Promise<boolean>;
    setReady: (submissionID: number, ready: boolean) => Promise<boolean>;
}

interface IResultsState {
    selectedSubmission?: IStudentLab;
    selectedStudent?: IStudentLabsForCourse;
    allSubmissions: IStudentLabsForCourse[];
}

export class Results extends React.Component<IResultsProp, IResultsState> {

    constructor(props: IResultsProp) {
        super(props);

        const currentStudent = this.props.allCourseSubmissions.length > 0 ? this.props.allCourseSubmissions[0] : null;
        const courseAssignments = currentStudent ? currentStudent.course.getAssignmentsList() : null;
        if (currentStudent && courseAssignments && courseAssignments.length > 0) {
            this.state = {
                // Only using the first student to fetch assignments.
                selectedSubmission: currentStudent.labs[0],
                allSubmissions: sortByScore(this.props.allCourseSubmissions, this.props.assignments, false),
            };
        } else {
            this.state = {
                selectedSubmission: undefined,
                allSubmissions: sortByScore(this.props.allCourseSubmissions, this.props.assignments, false),
            };
        }
    }

    public render() {
        let studentLab: JSX.Element | null = null;
        const currentStudents = this.props.allCourseSubmissions.length > 0 ? this.props.allCourseSubmissions : null;
        const currentLab = this.state.selectedSubmission;
        const currentSubmission = this.state.selectedSubmission?.submission;
        if (currentStudents
            && this.state.selectedSubmission && this.state.selectedStudent
            && !this.state.selectedSubmission.assignment.getIsgrouplab()
        ) {
            studentLab = <StudentLab
                studentSubmission={this.state.selectedSubmission}
                courseURL={this.props.courseURL}
                student={this.state.selectedStudent.enrollment.getUser() ?? new User()}
                teacherPageView={true}
                courseCreatorView={this.props.courseCreatorView}
                showApprove={true}
                slipdays={this.state.selectedSubmission.submission ? getSlipDays(this.props.allCourseSubmissions, this.state.selectedSubmission.submission, false) : 0}
                getReviewers={this.props.getReviewers}
                setApproved={async (submissionID: number, status: Submission.Status) => {
                    if (currentLab && currentSubmission) {
                        const ans = await this.props.setApproved(submissionID, status);
                        if (ans) {
                            currentSubmission.status = status;
                            this.setState({
                                selectedSubmission: currentLab,
                            });
                            // TODO: make sure the state is getting properly updated here
                        }
                    }
                }}
                setReady={async (submissionID: number, ready: boolean) => {
                    if (currentLab && currentSubmission) {
                        const ans = await this.props.setReady(submissionID, ready);
                        if (ans) {
                            currentSubmission.feedbackReady = ready;
                            this.setState({
                                selectedSubmission: currentLab,
                            });
                        }
                    }
                }}
                onRebuildClick={
                    async () => {
                        if (this.state.selectedSubmission && this.state.selectedSubmission.submission) {
                            const ans = await this.props.onRebuildClick(this.state.selectedSubmission.assignment.getId(), this.state.selectedSubmission.submission.id);
                            if (ans) {
                                this.state.selectedSubmission.submission = ans;
                                return true;
                            }
                        }
                        return false;
                    }
                }
                onApproveClick={ async (approve: boolean) => {
                    if (this.state.selectedSubmission && this.state.selectedSubmission.submission) {
                        const ans = await this.props.onApproveClick(this.state.selectedSubmission.submission.id, approve);
                        if (ans) {
                            this.state.selectedSubmission.submission.approved = approve;
                        }
                    }
                }}
            />;
        }

        return (
            <div>
                <h1>Result: {this.props.course.getName()}</h1>
                <Row>
                    <div key="resultshead" className="col-lg6 col-md-6 col-sm-12">
                        <Search className="input-group"
                            placeholder="Search for students"
                            onChange={(query) => this.handleSearch(query)}
                        />
                        <DynamicTable header={this.getResultHeader()}
                            data={this.state.allSubmissions}
                            selector={(item: IStudentLabsForCourse) => this.getResultSelector(item)}
                        />
                    </div>
                    <div key="resultsbody" className="col-lg-6 col-md-6 col-sm-12">
                        {studentLab}
                    </div>
                </Row>
            </div>
        );
    }

    private getResultHeader(): string[] {
        let headers: string[] = ["Name"];
        headers = headers.concat(this.props.assignments.filter((e) => !e.getIsgrouplab()).map((e) => e.getName()));
        return headers;
    }

    private getResultSelector(student: IStudentLabsForCourse): (string | JSX.Element | ICellElement)[] {
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
                        {e.submission ? (e.submission.score + "%") : "N/A"}</a>,
                    className: cellCss,
                };
                return iCell;
            }));
        return selector;
    }

    private handleOnclick(item: IStudentLab, student: IStudentLabsForCourse): void {
        this.setState({
            selectedSubmission: item,
            selectedStudent: student,
        });
    }

    private handleSearch(query: string): void {
        this.setState({
            allSubmissions: searchForLabs(this.props.allCourseSubmissions, query),
        });
    }
}
