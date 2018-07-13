import * as React from "react";
import { ICoursesWithAssignments, IUserCourse, IGroupCourse, ICourseLinkAssignment, IStudentSubmission } from "../../models";
import { ProgressBar } from "../progressbar/ProgressBar";

interface ISingleCourseOverviewProps {
    courseAndLabs: IUserCourse;
    groupAndLabs: IGroupCourse;
    onLabClick: (courseId: number, labId: number) => void;
}

class SingleCourseOverview extends React.Component<ISingleCourseOverviewProps, any> {
    private buildInfo(studentLabs: IStudentSubmission[], groupLabs: IStudentSubmission[]): IStudentSubmission[] | null {
        let labAndGrouplabs: IStudentSubmission[] = [];
        if (studentLabs.length != groupLabs.length) {
            return null;
        }
        for (var labCounter = 0; labCounter < studentLabs.length; labCounter++) {
            if (!studentLabs[labCounter].assignment.isgrouplab) {
                labAndGrouplabs.push(studentLabs[labCounter]);
            } else {
                labAndGrouplabs.push(groupLabs[labCounter]);
            }
        }

        return labAndGrouplabs;
    }

    public render() {
        var submissionArray = this.buildInfo(this.props.courseAndLabs.assignments, this.props.groupAndLabs.assignments);

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
                            {submission.assignment.deadline.toDateString()} <br />
                            {submission.assignment.deadline.toLocaleTimeString("en-GB")}
                        </span>
                    </div>
                </div>;
            }
            return (
                <li key={k} className="list-group-item clickable"
                    onClick={() => this.props.onLabClick(submission.assignment.courseid, submission.assignment.id)}>
                    <strong>{submission.assignment.name}</strong>
                    {submissionInfo}
                </li>);
        });
        return (
            <div>
                <h1>{this.props.courseAndLabs.course.name}</h1>
                <div>
                    <ul className="list-group">
                        {labs}
                    </ul>
                </div>
            </div>
        );
    }
}
export { SingleCourseOverview };
