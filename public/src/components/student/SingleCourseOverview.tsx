import * as React from "react";
import { ICoursesWithAssignments, IUserCourse } from "../../models";
import { ProgressBar } from "../progressbar/ProgressBar";

interface ISingleCourseOverviewProps {
    courseAndLabs: IUserCourse;
}

class SingleCourseOverview extends React.Component<ISingleCourseOverviewProps, any> {
    public render() {
        const labs: JSX.Element[] = this.props.courseAndLabs.assignments.map((v, k) => {
            return (
                <li key={k} className="list-group-item">
                    <strong>{v.assignment.name}</strong>
                    <ProgressBar progress={Math.floor((Math.random() * 100) + 1)} />
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
