import * as React from "react";
import { Assignment, Course, Review, User } from '../../../proto/ag_pb';
import { IStudentLabsForCourse, ISubmission } from '../../models';
import { ReviewPage } from '../../components/manual-grading/Review';
import { Search } from "../../components";
import { searchForLabs } from '../../componentHelper';

interface GradingViewProps {
    course: Course;
    courseURL: string;
    assignments: Assignment[];
    students: IStudentLabsForCourse[];
    curUser: User;
    addReview: (review: Review) => Promise<Review | null>;
    updateReview: (review: Review) => Promise<boolean>;

}

interface GradingViewState {
    selectedStudents: IStudentLabsForCourse[];
    selectedAssignment: Assignment;
    errorMessage: string;
}

export class GradingView extends React.Component<GradingViewProps, GradingViewState> {
    constructor(props: GradingViewProps) {
        super(props);
        this.state = {
            selectedStudents: this.props.students,
            selectedAssignment: this.props.assignments[0], // TODO: test on courses with no assignments
            errorMessage: "",
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
                 <select className="form-control">
                 {this.props.assignments.map((a, i) => <option
                            key={i}
                            className={i === 0 ? "active" : ""}
                            onClick={() => this.toggleAssignment(a)}
                        >{a.getName()}</option>)}
                    </select>
                    </div>
            </div>

            <div className="row"><div className="col-md-12">
                    <ul className="list-group">
                        {this.state.selectedStudents.map((s, i) =>
                            <li key={i} className="list-group-item li-review"><ReviewPage
                                key={"r" + i}
                                assignment={this.state.selectedAssignment}
                                submission={this.selectSubmission(s)}
                                authorName={s.enrollment.getUser()?.getName() ?? "Name not found"}
                                authorLogin={s.enrollment.getUser()?.getLogin() ?? "Login not found"}
                                courseURL={this.props.courseURL}
                                reviewerID={this.props.curUser.getId()}
                                addReview={this.props.addReview}
                                updateReview={this.props.updateReview}
                                studentNumber={this.props.students.indexOf(s) + 1}
                             /></li>
                        )}
                    </ul>
            </div></div>

        </div>
    }

    private selectSubmission(s: IStudentLabsForCourse): ISubmission | undefined {
        let lab: ISubmission | undefined;
        s.labs.forEach(l => {
            if (l.assignment.getId() === this.state.selectedAssignment.getId()) {
                lab = l.submission;
            }
        });
        return lab;
    }

    private toggleAssignment(a: Assignment) {
        this.setState({
            selectedAssignment: a,
        })
    }

    private handleSearch(query: string) {
        this.setState({
            selectedStudents: searchForLabs(this.props.students, query),
        })
    }

}