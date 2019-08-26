import * as React from "react";
import { IAssignmentLink, IStudentSubmission } from "../../models";
import { ProgressBar } from "../progressbar/ProgressBar";

interface ISingleCourseOverviewProps {
    courseAndLabs: IAssignmentLink;
    groupAndLabs?: IAssignmentLink;
    onLabClick: (courseId: number, labId: number) => void;
    onGroupLabClick: (courseId: number, labId: number) => void;
}

export class SingleCourseOverview extends React.Component<ISingleCourseOverviewProps, any> {
    public render() {
        let groupLabs: IStudentSubmission[] = [];
        if (this.props.groupAndLabs !== undefined) {
            groupLabs = this.props.groupAndLabs.assignments;
        }
        let submissionArray = this.buildInfo(this.props.courseAndLabs.assignments, groupLabs);

        // Fallback if the length of grouplabs and userlabs is different.
        if (!submissionArray) {
            submissionArray = this.props.courseAndLabs.assignments;
        }

        const labs: JSX.Element[] = submissionArray.map((submission, k) => {
            let submissionInfo = <div>No submissions</div>;
            if (submission.latest) {
                submissionInfo = <div className="row">
                    <div className="col-md-6 col-lg-8">
                        <ProgressBar progress={submission.latest.score} />
                    </div>
                    <div className="col-md-3 col-lg-2" >
                        <span className="text-success"> Passed: {submission.latest.passedTests} </span>
                        <span className="text-danger"> Failed: {submission.latest.failedTests} </span>
                    </div>
                    <div className="col-md-3 col-lg-2">
                        Deadline:
                        <span style={{ display: "inline-block", verticalAlign: "top", paddingLeft: "10px" }}>
                            {submission.assignment.getDeadline()}
                        </span>
                    </div>
                </div>;
            }
            return (
                <li key={k} className="list-group-item clickable"
                    // Testing if the onClick handler should be for studentlab or grouplab.
                    onClick={() => {
                        const courseId = submission.assignment.getCourseid();
                        const assignmentId = submission.assignment.getId();
                        if (!submission.assignment.getIsgrouplab()) {
                            return this.props.onLabClick(courseId, assignmentId);
                        } else {
                            return this.props.onGroupLabClick(courseId, assignmentId);
                        }
                    }}>
                    <strong>{submission.assignment.getName()}</strong>
                    {submissionInfo}
                </li >);
        });
        return (
            <div>
                <h1>{this.props.courseAndLabs.course.getName()}</h1>
                <div>
                    <ul className="list-group">
                        {labs}
                    </ul>
                </div>
            </div >
        );
    }
    private buildInfo(studentLabs: IStudentSubmission[], groupLabs: IStudentSubmission[]): IStudentSubmission[] | null {
        const labAndGrouplabs: IStudentSubmission[] = [];
        if (studentLabs.length !== groupLabs.length) {
            return null;
        }
        for (let labCounter = 0; labCounter < studentLabs.length; labCounter++) {
            if (!studentLabs[labCounter].assignment.getIsgrouplab()) {
                labAndGrouplabs.push(studentLabs[labCounter]);
            } else {
                labAndGrouplabs.push(groupLabs[labCounter]);
            }
        }

        return labAndGrouplabs;
    }
}
