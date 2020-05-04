import * as React from "react";
import { Course, GradingBenchmark, GradingCriterion, Assignment, User, Submission } from '../../../proto/ag_pb';
import { IStudentLabsForCourse, IReview, ISubmission } from '../../models';
import { Review } from '../../components/manual-grading/Review';


interface GradingViewProps {
    course: Course;
    assignments: Assignment[];
    students: IStudentLabsForCourse[];
    curUser: User;
    addReview: (review: IReview) => Promise<boolean>;
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
                student.labs.map((l, i) => <Review
                    key={"st" + i}
                    assignment={this.getAssignment(l.assignment)}
                    submission={l.submission}
                    review={this.selectReview(l.submission)}
                    authorName={l.authorName}
                    reviewerID={this.props.curUser.getId()}
                    addReview={(r: IReview) => this.props.addReview(r)}
                    updateReview={(r: IReview) => this.props.updateReview(r)}
                />)
            }</div>
        }
        return <div>No submissions yet from {this.state.selectedStudent?.enrollment.getUser()?.getName()}</div> // TODO: render empty view (some useful info on grading for TAs), i.e. list of active assignments for the course?
    }

    private getAssignment(a: Assignment): Assignment {
        console.log("Looking for assignment: " + a.toString());
        console.log("Found: " + this.props.assignments.find(item => item.getId() === a.getId())?.toString());
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
        return <div className="student-div"><ul className=" student-nav nav nav-stacked span2">
              {this.props.students.map((s, i) => <li
                key={"m" + i}
                className={this.setSelected(s)}
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