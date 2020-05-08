import * as React from "react";
import { Course, Assignment, Review, User } from '../../../proto/ag_pb';
import { IStudentLabsForCourse, ISubmission } from '../../models';
import { ReviewPage } from '../../components/manual-grading/Review';


interface GradingViewProps {
    course: Course;
    assignments: Assignment[];
    students: IStudentLabsForCourse[];
    curUser: User;
    addReview: (review: Review) => Promise<Review | null>;
    updateReview: (review: Review) => Promise<boolean>;

}

interface GradingViewState {
    selectedStudent: IStudentLabsForCourse | null;
    openState: boolean;
}

export class GradingView extends React.Component<GradingViewProps, GradingViewState> {
    constructor(props: GradingViewProps) {
        super(props);
        this.state = {
            selectedStudent: null,
            openState: true,
        }
    }

    public render() {
        return <div className="row grading-view">
            {this.renderStudentList()}{this.renderReview()}
        </div>
    }

    private renderReview(): JSX.Element {
        const student = this.state.selectedStudent;
        if (student && student.labs.length > 0) {
            return <div className="f-view">
                <h2 className="a-header">{student.labs[0].authorName}</h2>
                {
                student.labs.map((l, i) => <ReviewPage
                    key={"st" + i}
                    assignment={this.getAssignment(l.assignment)}
                    submission={l.submission}
                    authorName={l.authorName}
                    reviewerID={this.props.curUser.getId()}
                    review={this.selectReview(l.submission)}
                    open={this.state.openState}
                    setOpen={() => {
                        this.setState({
                            openState: true,
                        });
                    }}
                    addReview={async (r: Review) => {
                        if (l.submission) {
                            console.log("GradingView: adding a new review: " + r.toString());
                            const ans = await this.props.addReview(r);
                            if (ans) {
                                console.log("Review added successfully");
                                l.submission.reviews.push(ans);
                                return ans;
                            }
                        }
                        console.log("Failed to add review");
                        return null;
                    }}
                    updateReview={ async (r: Review) => {
                        if (l.submission) {
                            console.log("Grading view: updating review");
                            const ans = await this.props.updateReview(r);
                            if (ans) {
                                const ix = l.submission.reviews.findIndex(rw => rw.getId() === r.getId());
                                console.log("Review before update: " + l.submission.reviews[ix].toString());
                                l.submission.reviews[ix] = r;
                                console.log("Review after update: " + l.submission.reviews[ix].toString());
                                return true;
                            }
                        }
                        return false;
                    }}
                />)
            }</div>
        }
        return <div>{this.voidMessage()}</div>
    }

    private voidMessage(): string {
        return this.state.selectedStudent ? "No submissions yet from " + this.state.selectedStudent?.enrollment.getUser()?.getName() : "Select a course student for review";
    }

    private getAssignment(a: Assignment): Assignment {
        return this.props.assignments.find(item => item.getId() === a.getId()) ?? a;
    }

    private renderStudentList(): JSX.Element {
        return <div className="student-div"><ul className=" student-nav nav nav-pills nav-stacked flex-column">
              {this.props.students.map((s, i) => <li
                key={"m" + i}
                className={"nav-item" + this.setSelected(s)}
                onClick={() => {
                    this.setState({
                        selectedStudent: s,
                        openState: false,
                    })
                } }
              >{s.enrollment.getUser()?.getName() ?? "No name"}</li>)}
        </ul></div>;
    }

    // TODO: add style
    private setSelected(s: IStudentLabsForCourse): string {
        return this.state.selectedStudent === s ? "li-selected" : "";
    }

    private selectReview(s: ISubmission | undefined): Review | null {
        let rw: Review | null = null;
        if (s?.reviews) {
            s.reviews.forEach((r) => {
                if (r.getReviewerid() === this.props.curUser.getId()) {
                    console.log("Found an existing review by user: " + this.props.curUser.getLogin());
                    rw = r;
                }
            });
        }
        return rw;
    }
}