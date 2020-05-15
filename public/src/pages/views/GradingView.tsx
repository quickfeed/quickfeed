import * as React from "react";
import { Assignment, Course, Review, User } from '../../../proto/ag_pb';
import { IStudentLabsForCourse, ISubmission, IStudentLab } from '../../models';
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
    submissionsForAssignment: Map<User, IStudentLab>;
    errorMessage: string;
}

export class GradingView extends React.Component<GradingViewProps, GradingViewState> {
    constructor(props: GradingViewProps) {
        super(props);
        this.state = {
            selectedStudents: this.props.students,
            selectedAssignment: this.props.assignments[0] ?? new Assignment(), // TODO: test on courses with no assignments
            errorMessage: "",
            submissionsForAssignment: this.props.assignments[0] ? this.selectAllSubmissions(this.props.assignments[0]) : new Map<User, IStudentLab>(),
        }
    }

    public render() {
        if (this.props.assignments.length < 1) {
            return <div className="alert alert-info">No assignments for {this.props.course.getName()} </div>
        }
        const allStudents = Array.from(this.state.submissionsForAssignment.keys());

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
                        {allStudents.map((s, i) =>
                            <li key={i} className="list-group-item li-review"><ReviewPage
                                key={"r" + i}
                                assignment={this.state.selectedAssignment}
                                submission={this.state.submissionsForAssignment.get(s)?.submission}
                                authorName={s.getName() ?? "Name not found"}
                                authorLogin={s.getLogin() ?? "Login not found"}
                                courseURL={this.props.courseURL}
                                reviewerID={this.props.curUser.getId()}
                                addReview={this.props.addReview}
                                updateReview={this.props.updateReview}
                                studentNumber={allStudents.indexOf(s) + 1}
                             /></li>
                        )}
                    </ul>
            </div></div>

        </div>
    }

    private selectAllSubmissions(a?: Assignment): Map<User, IStudentLab> {
        const labMap = this.state.submissionsForAssignment;
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

    private toggleAssignment(a: Assignment) {
        console.log("Setting assignment " + a.getName());
        this.setState({
            selectedAssignment: a,
            submissionsForAssignment: this.selectAllSubmissions(a),
        });
    }

    private handleSearch(query: string) {
        this.setState({
            selectedStudents: searchForLabs(this.props.students, query),
        })
    }

}