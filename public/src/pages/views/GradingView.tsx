import * as React from "react";
import { Assignment, Course, Group, Review, User, Submission } from "../../../proto/ag_pb";
import { IAllSubmissionsForEnrollment, ISubmission, ISubmissionLink } from '../../models';
import { ReviewPage } from "../../components/manual-grading/Review";
import { Search } from "../../components";
import { searchForUsers, sortStudentsForRelease, totalScore } from '../../componentHelper';
import { Release } from "../../components/manual-grading/Release";

interface GradingViewProps {
    course: Course;
    courseURL: string;
    assignments: Assignment[];
    students: IAllSubmissionsForEnrollment[];
    curUser: User;
    releaseView: boolean;
    addReview: (review: Review) => Promise<Review | null>;
    updateReview: (review: Review) => Promise<boolean>;
    onUpdate: (submission: ISubmission) => Promise<boolean>;
    getReviewers: (submissionID: number) => Promise<User[]>;
    releaseAll: (assignmentID: number, score: number, release: boolean, approve: boolean) => Promise<boolean>;
}

interface GradingViewState {
    selectedStudent: User | undefined;
    selectedStudents: User[];
    selectedAssignment: Assignment;
    submissionsForAssignment: Map<User, ISubmissionLink>;
    submissionsForGroupAssignment: Map<Group, ISubmissionLink>;
    alert: string;
    scoreLimit: number;
}

export class GradingView extends React.Component<GradingViewProps, GradingViewState> {
    constructor(props: GradingViewProps) {
        super(props);
        this.state = {
            selectedStudent: undefined,
            selectedStudents: [],
            selectedAssignment: this.props.assignments[0] ?? new Assignment(), // TODO: test on courses with no assignments
            alert: "",
            submissionsForAssignment: this.props.assignments[0] ? this.selectAllSubmissions(this.props.assignments[0]) : new Map<User, ISubmissionLink>(),
            submissionsForGroupAssignment: new Map<Group, ISubmissionLink>(),
            scoreLimit: 0,
        }
    }

    public render() {
        if (this.props.assignments.length < 1) {
            return <div className="alert alert-info">No assignments for {this.props.course.getName()} </div>
        }
        return <div className="grading-view">
            <div className="row"><h1>Review submissions for {this.props.course.getName()}</h1></div>

            <div className="row"><div className="col-md-8"><Search className="input-group"
                    placeholder="Search for students"
                    onChange={(query) => this.handleSearch(query)}
                /></div>
                 <div className="form-group col-md-4">
                 <select className="form-control" onChange={(e) => this.toggleAssignment(e.target.value)}>
                 {this.props.assignments.map((a, i) => <option
                            key={i}
                            value={a.getId()}
                       >{a.getName()}</option>)}Select assignment
                    </select>
                    </div>
            </div>

            {this.renderAlert()}

            {this.props.releaseView ? this.renderReleaseRow() : null}

            <div className="row">
                {this.props.releaseView ? this.renderReleaseList() : this.renderReviewList()}
            </div>

        </div>
    }

    public componentDidMount() {
        this.setState((state) => ({
            selectedStudents: this.props.releaseView ? sortStudentsForRelease(state.submissionsForAssignment, state.selectedAssignment.getReviewers()) : this.selectAllStudents(),
            submissionsForAssignment: this.selectAllSubmissions(state.selectedAssignment),
        }));
    }

    private renderAlert(): JSX.Element | null {
        return this.state.alert === "" ? null : <div className="row"><div className="alert alert-warning">{ this.state.alert }</div></div>
    }

    private renderReleaseRow(): JSX.Element {
        return <div className="row"><div className="col-md-12">
            <div className="input-group">
                <span className="input-group-addon">Set minimal score:</span>
                <input
                    className="form-control m-input"
                    type="number"
                    min="0"
                    max="100"
                    defaultValue={this.state.selectedAssignment.getScorelimit()}
                    value={this.state.scoreLimit}
                    onChange={(e) => {
                        this.setState({
                            scoreLimit: parseInt(e.target.value, 10),
                        });
                    }}
                />
                <div className="input-group-btn">
                    <button className="btn btn-default"
                        onClick={async () => {
                            if (this.state.scoreLimit < 1) {
                                this.setState({
                                  alert: "Minimal score for approving is not set",
                            });
                        } else {
                            const ans = this.props.releaseAll(this.state.selectedAssignment.getId(), this.state.scoreLimit, false, true);
                            if (ans) {
                                this.setState((state) => ({
                                    selectedStudents: this.props.releaseView ? sortStudentsForRelease(state.submissionsForAssignment, state.selectedAssignment.getReviewers()) : this.selectAllStudents(),
                                    submissionsForAssignment: this.selectAllSubmissions(state.selectedAssignment),
                                    alert: "",
                                }));
                            } else {
                                this.setState({
                                    alert: "Failed to approve submissions",
                                });
                            }
                        }}

                    }
                    >Approve all</button>
                </div>
                <div className="input-group-btn">
                <button className="btn btn-default"
                        onClick={() => {
                            if (this.state.scoreLimit < 1) {
                                this.setState({
                                    alert: "Minimal score for releasing is not set",
                                });
                            } else {
                                const ans = this.props.releaseAll(this.state.selectedAssignment.getId(), this.state.scoreLimit, true, false);
                                if (ans) {
                                    this.setState((state) => ({
                                        selectedStudents: this.props.releaseView ? sortStudentsForRelease(state.submissionsForAssignment, state.selectedAssignment.getReviewers()) : this.selectAllStudents(),
                                        submissionsForAssignment: this.selectAllSubmissions(state.selectedAssignment),
                                        alert: "",
                                    }));
                                } else {
                                    this.setState({
                                        alert: "Failed to release submissions",
                                    });
                                }
                            }}
                        }
                    >Release all</button>
                </div>
            </div>
        </div></div>
    }

    private renderReleaseList(): JSX.Element {
        const allCourseStudents = this.selectAllStudents();
        return <div className="col-md-12">
            <ul className="list-group">
                {
                    this.state.selectedStudents.map((s, i) =>
                        <li key={i} onClick={() => this.setState({selectedStudent: s})} className="list-group-item li-review"><Release
                            key={"f" + i}
                            teacherView={true}
                            assignment={this.state.selectedAssignment}
                            submission={this.state.submissionsForAssignment.get(s)?.submission}
                            authorName={s.getName()}
                            authorLogin={s.getLogin()}
                            courseURL={this.props.courseURL}
                            isSelected={this.state.selectedStudent === s}
                            setGrade={async (status: Submission.Status, approved: boolean) => {
                                const current = this.state.submissionsForAssignment.get(s);
                                if (current && current.submission) {
                                    current.submission.status = status;
                                    current.submission.approved = approved;
                                    current.submission.score = totalScore(current.submission.reviews);
                                    return this.props.onUpdate(current.submission);
                                }
                                return false;
                            }}
                            release={async (release: boolean) => {
                                const current = this.state.submissionsForAssignment.get(s)?.submission;
                                if (current) {
                                    current.released = release;
                                    current.score = totalScore(current.reviews);
                                    const ans = this.props.onUpdate(current);
                                    if (ans) return true;
                                    current.released = !release;
                                    return false;
                                }
                            }}
                            getReviewers={this.props.getReviewers}
                            studentNumber={allCourseStudents.indexOf(s) + 1}
                        /></li>
                    )
                }
            </ul>
        </div>
    }

    private renderReviewList(): JSX.Element {
        const allCourseStudents = this.selectAllStudents();
        return <div className="col-md-12">
        <ul className="list-group">
            {this.state.selectedStudents.map((s, i) =>
                <li key={i} onClick={() => this.setState({selectedStudent: s})} className="list-group-item li-review"><ReviewPage
                    key={"r" + i}
                    assignment={this.state.selectedAssignment}
                    submission={this.state.submissionsForAssignment.get(s)?.submission}
                    authorName={s.getName() ?? "Name not found"}
                    authorLogin={s.getLogin() ?? "Login not found"}
                    courseURL={this.props.courseURL}
                    reviewerID={this.props.curUser.getId()}
                    addReview={async (review: Review) => {
                        const current = this.state.submissionsForAssignment.get(s);
                        if (current?.submission) {
                            const ans = await this.props.addReview(review);
                            if (ans) {
                                current.submission.reviews.push(ans);
                                return true;
                            }
                        }
                        return false;
                    }}
                    updateReview={async (review: Review) => {
                        const current = this.state.submissionsForAssignment.get(s);
                        if (current?.submission) {
                            return this.props.updateReview(review);
                        }
                        return false;
                    }}
                    studentNumber={allCourseStudents.indexOf(s) + 1}
                    isSelected={this.state.selectedStudent === s}
                     /></li>
                )}
            </ul>
        </div>
    }

    private selectAllStudents(): User[] {
        const studentUsers: User[] = [];
        this.props.students.forEach(s => {
            studentUsers.push(s.enrollment.getUser() ?? new User());
        });
        return studentUsers;
    }

    private selectAllSubmissions(a?: Assignment): Map<User, ISubmissionLink> {
        const labMap = new Map<User, ISubmissionLink>();
        const current = a ?? this.state.selectedAssignment;
        this.props.students.forEach(s => {
            s.labs.forEach(l => {
                if (l.assignment.getId() === current.getId()) {
                    labMap.set(s.enrollment.getUser() ?? new User(), l);
                }
            });
        });
        return labMap;
    }

    private toggleAssignment(id: string) {
        const currentID = parseInt(id, 10);
        const current = this.props.assignments.find(item => item.getId() === currentID);
        if (current) {
            const submissionsList = this.selectAllSubmissions(current);
            this.setState({
                selectedStudent: undefined,
                selectedAssignment: current,
                submissionsForAssignment: submissionsList,
                selectedStudents: this.props.releaseView ? sortStudentsForRelease(submissionsList, current.getReviewers()) : this.selectAllStudents(),
            });
        }
    }

    private handleSearch(query: string) {
        this.setState({
            selectedStudents: searchForUsers(this.selectAllStudents(), query),
        });
    }

}