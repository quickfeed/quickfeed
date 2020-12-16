import * as React from "react";
import { Assignment, Course, Group, User, Submission } from "../../../proto/ag_pb";
import { IAllSubmissionsForEnrollment, ISubmission, ISubmissionLink } from "../../models";
import { Search } from "../../components";
import { mapAllSubmissions, sortStudentsForRelease, scoreFromReviews, selectFromSubmissionLinks, searchForUsers, searchForGroups, sortAssignmentsByOrder } from "../../componentHelper";
import { Release } from "../../components/manual-grading/Release";

interface ReleaseViewProps {
    course: Course;
    courseURL: string;
    assignments: Assignment[];
    students: IAllSubmissionsForEnrollment[];
    groups: IAllSubmissionsForEnrollment[];
    curUser: User;
    onUpdate: (submission: ISubmission) => Promise<boolean>;
    getReviewers: (submissionID: number) => Promise<User[]>;
    updateAll: (assignmentID: number, score: number, release: boolean, approve: boolean) => Promise<boolean>;
}

interface ReleaseViewState {
    allStudents: User[]; // immutable, only set once in constructor
    allGroups: Group[]; // immutable, only set once in constructor
    showStudents: User[];
    showGroups: Group[];
    selectedAssignment: Assignment | undefined;
    selectedStudent: User | undefined;
    selectedGroup: Group | undefined;
    submissionsForAssignment: Map<User, ISubmissionLink>; // recalculate on new assignment
    submissionsForGroupAssignment: Map<Group, ISubmissionLink>; // recalculate on new group assignment
    alert: string;
    scoreLimit: number;
}

export class ReleaseView extends React.Component<ReleaseViewProps, ReleaseViewState> {
    constructor(props: ReleaseViewProps) {
        super(props);
        const a = this.props.assignments[0];
        this.state = {
            selectedStudent: undefined,
            selectedGroup: undefined,
            allStudents: selectFromSubmissionLinks(props.students, false) as User[],
            allGroups: selectFromSubmissionLinks(props.groups, true) as Group[],
            showStudents: selectFromSubmissionLinks(props.students, false) as User[],
            showGroups: selectFromSubmissionLinks(props.groups, true) as Group[],
            selectedAssignment: a,
            alert: "",
            submissionsForAssignment: mapAllSubmissions(props.students, false, a) as Map<User, ISubmissionLink>,
            submissionsForGroupAssignment: mapAllSubmissions(props.groups, true, a) as Map<Group, ISubmissionLink>,
            scoreLimit: a ? a.getScorelimit() : 0,
        }
    }

    public render() {
        if (this.props.assignments.length < 1) {
            return <div className="row"><div className="alert alert-info col-md-12">No assignments for {this.props.course.getName()}. </div></div>;
        }
        return <div className="release-view">
            <div className="row"><h1>Release submissions for {this.props.course.getName()}</h1></div>

            <div className="row"><div className="col-md-8"><Search className="input-group"
                    placeholder="Search for students or groups"
                    onChange={(query) => this.handleSearch(query)}
                /></div>
                 <div className="form-group col-md-4">
                 <select className="form-control" onChange={(e) => this.toggleAssignment(e.target.value)}>
                 {sortAssignmentsByOrder(this.props.assignments).map((a, i) => <option
                            key={i}
                            value={a.getId()}
                       >{a.getName()}</option>)}Select assignment
                    </select>
                    </div>
            </div>

            {this.renderAlert()}

            {this.renderReleaseRow()}

            <div className="row">
                {this.renderReleaseList()}
            </div>

        </div>
    }

    private renderAlert(): JSX.Element | null {
        return this.state.alert === "" ? null : <div className="row"><div className="alert alert-warning col-md-12">{ this.state.alert }</div></div>
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
                    value={this.state.scoreLimit > 0 ? this.state.scoreLimit : ""}
                    onChange={(e) => {
                        this.setState({
                            scoreLimit: e.target.value !== "" ? parseInt(e.target.value, 10) : 0,
                        });
                    }}
                />
                <div className="input-group-btn">
                    <button className="btn btn-default"
                        onClick={() => {
                            if (this.state.selectedAssignment) {
                                if (confirm(
                                    `Warning! Are you sure you want to approve all submissions with score above ${this.state.scoreLimit}?`,
                                    )) {
                                        this.props.updateAll(this.state.selectedAssignment.getId(), this.state.scoreLimit, false, true);
                                }
                            }
                            this.setState({
                                alert: this.alertWhenMassReleasing(),
                            })
                        }}
                    >Approve all</button>
                </div>
                <div className="input-group-btn">
                <button className="btn btn-default"
                        onClick={() => {
                            if (this.state.selectedAssignment) {
                                if (confirm(
                                    `Warning! Are you sure you want to release reviews for allsubmissions with score above ${this.state.scoreLimit}?`,
                                    )) {
                                        this.props.updateAll(this.state.selectedAssignment.getId(), this.state.scoreLimit, true, false);
                                }
                            }
                            this.setState({
                                alert: this.alertWhenMassReleasing(),
                            });
                        }}
                    >Release all</button>
                </div>
            </div>
        </div></div>
    }

    private alertWhenMassReleasing(): string {
        if (this.state.scoreLimit > 100) return "Score cannot be above 100";
        if (!this.state.selectedAssignment) return "No assignment is selected";
        return "";
    }

    private renderReleaseList(): JSX.Element {
        const a = this.state.selectedAssignment;
        if (!a) {
            return <div className="alert alert-dark col-md-12">Please select an assignment.</div>;
        }
        if (a.getIsgrouplab()) {
            const sortedCourseGroups = sortStudentsForRelease<Group>(this.state.showGroups, this.state.submissionsForGroupAssignment, a.getReviewers());
            return <div className="col-md-12">
                <ul className="list-group">{
                    sortedCourseGroups.map((grp, i) =>
                        <li key={i} onClick={() => this.setState({selectedGroup: grp})} className="list-group-item li-review"><Release
                            key={"fg" + i}
                            teacherView={true}
                            userIsCourseCreator={this.props.course.getCoursecreatorid() === this.props.curUser.getId()}
                            assignment={a}
                            submission={this.state.submissionsForGroupAssignment.get(grp)?.submission}
                            authorName={grp.getName()}
                            authorLogin={grp.getName()}
                            courseURL={this.props.courseURL}
                            isSelected={this.state.selectedGroup === grp}
                            setGrade={async (status: Submission.Status, approved: boolean) => {
                                const current = this.state.submissionsForGroupAssignment.get(grp);
                                if (current && current.submission) {
                                    current.submission.status = status;
                                    current.submission.score = scoreFromReviews(current.submission.reviews);
                                    return this.props.onUpdate(current.submission);
                                }
                                return false;
                            }}
                            release={async (release: boolean) => {
                                const current = this.state.submissionsForGroupAssignment.get(grp)?.submission;
                                if (current) {
                                    current.released = release;
                                    current.score = scoreFromReviews(current.reviews);
                                    const ans = this.props.onUpdate(current);
                                    if (ans) return true;
                                    current.released = !release;
                                    return false;
                                }
                            }}
                            getReviewers={this.props.getReviewers}
                            studentNumber={this.state.allGroups.indexOf(grp) + 1}
                        /></li>)
                }</ul>
            </div>

        }
        const sortedCourseStudents = sortStudentsForRelease<User>(this.state.showStudents, this.state.submissionsForAssignment, a.getReviewers());
        return <div className="col-md-12">
            <ul className="list-group">
                {
                    sortedCourseStudents.map((s, i) =>
                        <li key={i} onClick={() => this.setState({selectedStudent: s})} className="list-group-item li-review"><Release
                            key={"fs" + i}
                            userIsCourseCreator={this.props.course.getCoursecreatorid() === this.props.curUser.getId()}
                            teacherView={true}
                            assignment={a}
                            submission={this.state.submissionsForAssignment.get(s)?.submission}
                            authorName={s.getName()}
                            authorLogin={s.getLogin()}
                            courseURL={this.props.courseURL}
                            isSelected={this.state.selectedStudent === s}
                            setGrade={async (status: Submission.Status, approved: boolean) => {
                                const current = this.state.submissionsForAssignment.get(s);
                                if (current && current.submission) {
                                    current.submission.status = status;
                                    current.submission.score = scoreFromReviews(current.submission.reviews);
                                    return this.props.onUpdate(current.submission);
                                }
                                return false;
                            }}
                            release={async (release: boolean) => {
                                const current = this.state.submissionsForAssignment.get(s)?.submission;
                                if (current) {
                                    current.released = release;
                                    current.score = scoreFromReviews(current.reviews);
                                    const ans = this.props.onUpdate(current);
                                    if (ans) return true;
                                    current.released = !release;
                                    return false;
                                }
                            }}
                            getReviewers={this.props.getReviewers}
                            studentNumber={this.state.allStudents.indexOf(s) + 1}
                        /></li>)
                    }</ul>
        </div>
    }

    private toggleAssignment(id: string) {
        const currentID = parseInt(id, 10);
        const current = this.props.assignments.find(item => item.getId() === currentID);
        if (current) {
            this.setState({
                selectedStudent: undefined,
                selectedGroup: undefined,
                selectedAssignment: current,
                submissionsForAssignment: mapAllSubmissions(this.props.students, false, current) as Map<User, ISubmissionLink>,
                submissionsForGroupAssignment: mapAllSubmissions(this.props.groups, true, current) as Map<Group, ISubmissionLink>,
                scoreLimit: current.getScorelimit(),
            });
        }
    }

    private handleSearch(query: string) {
        const foundUsers = searchForUsers(sortStudentsForRelease<User>(this.state.allStudents, this.state.submissionsForAssignment, this.state.selectedAssignment?.getReviewers() ?? 0), query);
        const foundGroups = searchForGroups(sortStudentsForRelease<Group>(this.state.allGroups, this.state.submissionsForGroupAssignment, this.state.selectedAssignment?.getReviewers() ?? 0), query);
        this.setState((state) => ({
            showStudents: foundUsers,
            showGroups: foundGroups,
        }));
    }

}