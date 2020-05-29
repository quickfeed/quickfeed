import * as React from "react";
import { Assignment, Course, Group, Review, User, Submission } from "../../../proto/ag_pb";
import { IAllSubmissionsForEnrollment, ISubmission, ISubmissionLink } from '../../models';
import { ReviewPage } from "../../components/manual-grading/Review";
import { Search } from "../../components";
import { searchForUsers, sortStudentsForRelease, totalScore, selectFromSubmissionLinks, mapAllSubmissions } from '../../componentHelper';

interface FeedbackViewProps {
    course: Course;
    courseURL: string;
    assignments: Assignment[];
    students: IAllSubmissionsForEnrollment[];
    groups: IAllSubmissionsForEnrollment[];
    curUser: User;
    addReview: (review: Review) => Promise<Review | null>;
    updateReview: (review: Review) => Promise<boolean>;
}

interface FeedbackViewState {
    allStudents: User[]; // immutable, only set once in constructor
    allGroups: Group[]; // immutable, only set once in constructor
    selectedAssignment: Assignment;
    selectedStudent: User | undefined;
    selectedGroup: Group | undefined;
    submissionsForAssignment: Map<User, ISubmissionLink>; // recalculate on new assignment
    submissionsForGroupAssignment: Map<Group, ISubmissionLink>; // recalculate on new group assignment
    alert: string;
}

export class FeedbackView extends React.Component<FeedbackViewProps, FeedbackViewState> {

    constructor(props: FeedbackViewProps) {
        super(props);
        const a = this.props.assignments[0];
        this.state = {
            selectedStudent: undefined,
            selectedGroup: undefined,
            allStudents: selectFromSubmissionLinks(props.students, false) as User[],
            allGroups: selectFromSubmissionLinks(props.groups, true) as Group[],
            selectedAssignment: a,
            alert: "",
            submissionsForAssignment: mapAllSubmissions(props.students, false, a) as Map<User, ISubmissionLink>,
            submissionsForGroupAssignment: mapAllSubmissions(props.groups, true, a) as Map<Group, ISubmissionLink>,
        }
    }


/*
    private renderReviewList(): JSX.Element {
        const allCourseStudents = this.selectAllStudents();
        return <div className="col-md-12">
        <ul className="list-group">
            {this.state.allStudents.map((s, i) =>
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
    }*/


}