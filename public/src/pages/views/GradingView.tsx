import * as React from "react";
import { Assignment, Course, Review, User, Submission } from "../../../proto/ag_pb";
import { IAllSubmissionsForEnrollment, ISubmission, ISubmissionLink } from "../../models";
import { ReviewPage } from "../../components/manual-grading/Review";
import { Search } from "../../components";
import { searchForUsers, sortStudentsForRelease } from '../../componentHelper';
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
}

interface GradingViewState {

    selectedStudents: User[];
    selectedAssignment: Assignment;
    submissionsForAssignment: Map<User, ISubmissionLink>;
    errorMessage: string;
    allClosed: boolean;
    scoreLimit: number;
}

export class GradingView extends React.Component<GradingViewProps, GradingViewState> {
    constructor(props: GradingViewProps) {
        super(props);
        this.state = {
            selectedStudents: this.selectAllStudents(),
            selectedAssignment: this.props.assignments[0] ?? new Assignment(), // TODO: test on courses with no assignments
            errorMessage: "",
            submissionsForAssignment: this.props.assignments[0] ? this.selectAllSubmissions(this.props.assignments[0]) : new Map<User, ISubmissionLink>(),
            allClosed: true,
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

            {this.props.releaseView ? this.renderReleaseRow() : null}

            <div className="row">
                {this.props.releaseView ? this.renderReleaseList() : this.renderReviewList()}
            </div>

        </div>
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
                    value={this.state.scoreLimit}
                    onChange={(e) => {
                        this.setState({
                            scoreLimit: parseInt(e.target.value, 10),
                        });
                    }}
                />
                <div className="input-group-btn">
                    <button className="btn btn-default"
                        onClick={() => {
                          if (this.state.scoreLimit < 1) {
                              this.setState({
                                  errorMessage: "Minimal score for approving is not set",
                              });
                          } else {
                            console.log("Approving all submissions with score >= " + this.state.scoreLimit)}
                          }

                        }
                    >Approve all</button>
                </div>
                <div className="input-group-btn">
                <button className="btn btn-default"
                        onClick={() => {
                            if (this.state.scoreLimit < 1) {
                                this.setState({
                                    errorMessage: "Minimal score for releasing is not set",
                                });
                            } else {
                              console.log("Releasing reviews for all submission with score >= " + this.state.scoreLimit)}
                            }

                        }
                    >Release all</button>
                </div>
            </div>
        </div></div>
    }

    private renderReleaseList(): JSX.Element {
        const sortedStudents = sortStudentsForRelease(this.state.submissionsForAssignment, this.state.selectedAssignment.getReviewers());
        return <div className="col-md-12">
            <ul className="list-group">
                {
                    sortedStudents.map((s, i) =>
                        <li key={i} className="list-group-item li-review"><Release
                            key={"f" + i}
                            teacherView={true}
                            assignment={this.state.selectedAssignment}
                            submission={this.state.submissionsForAssignment.get(s)?.submission}
                            authorName={s.getName()}
                            authorLogin={s.getLogin()}
                            courseURL={this.props.courseURL}
                            allClosed={this.state.allClosed}
                            setGrade={async (status: Submission.Status, approved: boolean) => {
                                const current = this.state.submissionsForAssignment.get(s);
                                if (current && current.submission) {
                                    current.submission.status = status;
                                    current.submission.approved = approved;
                                    return this.props.onUpdate(current.submission);
                                }
                                return false;
                            }}
                            release={async (release: boolean) => {
                                const current = this.state.submissionsForAssignment.get(s)?.submission;
                                if (current) {
                                    current.released = release;
                                    const ans = this.props.onUpdate(current);
                                    if (ans) return true;
                                    current.released = !release;
                                    return false;
                                }
                            }}
                            getReviewers={this.props.getReviewers}
                            studentNumber={this.state.selectedStudents.indexOf(s) + 1}
                            toggleCloseAll={() => this.toggleClosed()}
                        /></li>
                    )
                }
            </ul>
        </div>
    }

    private toggleClosed() {
        this.setState({
            allClosed: !this.state.allClosed,
            selectedStudents: this.selectAllStudents(),
            submissionsForAssignment: this.state.selectedAssignment ? this.selectAllSubmissions(this.state.selectedAssignment) : this.selectAllSubmissions(this.props.assignments[0]),
            errorMessage: "",
        });
    }

    private renderReviewList(): JSX.Element {
        return <div className="col-md-12">
        <ul className="list-group">
            {this.state.selectedStudents.map((s, i) =>
                <li key={i} className="list-group-item li-review"><ReviewPage
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
                    studentNumber={this.state.selectedStudents.indexOf(s) + 1}
                    allClosed={this.state.allClosed}
                    toggleCloseAll={() => this.toggleClosed()}
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
            this.setState({
                selectedAssignment: current,
                submissionsForAssignment: this.selectAllSubmissions(current),
            });
        }
    }

    private handleSearch(query: string) {
        this.setState({
            selectedStudents: searchForUsers(this.selectAllStudents(), query),
        });
    }

}