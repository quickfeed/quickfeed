import * as React from "react";
import { ICoursesWithAssignments } from "../../models";
import { ProgressBar } from "../progressbar/ProgressBar";

interface ISingleCourseOverviewProps {
    courseAndLabs: ICoursesWithAssignments;
}

class SingleCourseOverview extends React.Component<ISingleCourseOverviewProps, any> {
    public render() {
        const labs: JSX.Element[] = this.props.courseAndLabs.labs.map((v, k) => {
            return (
                <li key={k} className="list-group-item">
                    <strong>{v.name}</strong>
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
