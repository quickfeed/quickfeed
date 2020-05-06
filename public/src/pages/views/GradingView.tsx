import * as React from "react";
import { Course, Assignment, User } from '../../../proto/ag_pb';
import { IStudentLabsForCourse, IReview, ISubmission } from '../../models';
import { ReviewPage } from '../../components/manual-grading/Review';


interface GradingViewProps {
    course: Course;
    assignments: Assignment[];
    students: IStudentLabsForCourse[];
    curUser: User;
    addReview: (review: IReview) => Promise<IReview | null>;
    updateReview: (review: IReview) => Promise<boolean>;

}

interface GradingViewState {
    selectedStudent: IStudentLabsForCourse | null;
}

export class GradingView extends React.Component<GradingViewProps, GradingViewState> {
    constructor(props: GradingViewProps) {
        super(props);
        this.state = {
            selectedStudent: null,
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
                    review={this.selectReview(l.submission)}
                    authorName={l.authorName}
                    reviewerID={this.props.curUser.getId()}
                    addReview={async (r: IReview) => {
                        if (l.submission) {
                            const ans = await this.props.addReview(r);
                            if (ans) {
                                l.submission.reviews.push(r);
                                return true;
                            }
                        }
                        return false;
                    }}
                    updateReview={ async (r: IReview) => {
                        if (l.submission) {
                            const ans = await this.props.updateReview(r);
                            if (ans) {
                                const ix = l.submission.reviews.findIndex(rw => rw.id === r.id);
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

    private selectReview(s: ISubmission | undefined): IReview | null {
        let review: IReview | null = null;
        s?.reviews.forEach((r) => {
            if (r.reviewerID === this.props.curUser.getId()) {
                review = r;
            }
        });
        return review;
    }

    private renderStudentList(): JSX.Element {
        return <div className="student-div"><ul className=" student-nav nav nav-pills nav-stacked flex-column">
              {this.props.students.map((s, i) => <li
                key={"m" + i}
                className={"nav-item" + this.setSelected(s)}
                onClick={() => {
                    this.setState({
                        selectedStudent: s,
                    })
                } }
              >{s.enrollment.getUser()?.getName() ?? "No name"}</li>)}
        </ul></div>;
    }

    // TODO: add style
    private setSelected(s: IStudentLabsForCourse): string {
        return this.state.selectedStudent === s ? "li-selected" : "";
    }
}