import * as React from "react";
import { ICoursesWithAssignments, IUserCourse } from "../../models";
import { ProgressBar } from "../progressbar/ProgressBar";

interface ISingleCourseOverviewProps {
    courseAndLabs: IUserCourse;
    onLabClick: (courseId: number, labId: number) => void;
}

class SingleCourseOverview extends React.Component<ISingleCourseOverviewProps, any> {
    public render() {
        const labs: JSX.Element[] = this.props.courseAndLabs.assignments.map((submission, k) => {
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
