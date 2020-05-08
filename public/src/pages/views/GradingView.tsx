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
                            const ans = await this.props.addReview(r);
                            if (ans) {
                                l.submission.reviews.push(ans);
                                return ans;
                            }
                        }
                        return null;
                    }}
                    updateReview={ async (r: Review) => {
                        if (l.submission) {
                            const ans = await this.props.updateReview(r);
                            if (ans) {
                                const ix = l.submission.reviews.findIndex(rw => rw.getId() === r.getId());
                                l.submission.reviews[ix] = r;
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
        return <div className="student-div"><ul className=" student-nav nav nav-pills nav-fill flex-column">
              {this.props.students.map((s, i) => <li
                key={"m" + i}
                className={"nav-item nav-link" + this.setSelected(s)}
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
        return this.state.selectedStudent === s ? "active" : "";
    }

    private selectReview(s: ISubmission | undefined): Review | null {
        let rw: Review | null = null;
        if (s?.reviews) {
            s.reviews.forEach((r) => {
                if (r.getReviewerid() === this.props.curUser.getId()) {
                    rw = r;
                }
            });
        }
        return rw;
    }
}