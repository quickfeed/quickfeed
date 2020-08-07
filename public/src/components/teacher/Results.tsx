import * as React from "react";
import { Assignment, Course, User, Submission } from '../../../proto/ag_pb';
import { DynamicTable, Row, Search, StudentLab } from "../../components";
import { IAllSubmissionsForEnrollment, ISubmissionLink, ISubmission } from '../../models';
import { ICellElement } from "../data/DynamicTable";
import { generateCellClass, sortByScore } from "./labHelper";
import { searchForLabs, userRepoLink, getSlipDays } from '../../componentHelper';

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
                onSubmissionRebuild={
                    async () => {
                        if (this.state.selectedSubmission && this.state.selectedSubmission.submission) {
                            const ans = await this.props.onSubmissionRebuild(this.state.selectedSubmission.assignment.getId(), this.state.selectedSubmission.submission.id);
                            if (ans) {
                                this.state.selectedSubmission.submission = ans;
                                return true;
                            }
                        }
                        return false;
                    }
                }
                onSubmissionStatusUpdate={ async (status: Submission.Status) => {
                    const current = this.state.selectedSubmission;
                    const selected = current?.submission;
                    if (selected) {
                        selected.status = status;
                        const ans = await this.props.onSubmissionStatusUpdate(selected);
                        if (ans) {
                            this.setState({
                                selectedSubmission: current,
                            })
                        }
                        return ans;
                    }
                    return false;
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
                        {e.submission ? (e.submission.score + "%") : "N/A"}</a>,
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
        });
    }

    private handleSearch(query: string): void {
        this.setState({
            allSubmissions: searchForLabs(this.props.allCourseSubmissions, query),
        });
    }
}
