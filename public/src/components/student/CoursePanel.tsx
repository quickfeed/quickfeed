import * as React from "react";

import { DynamicTable } from "../../components";

import { Assignment, Course } from "../../../proto/ag_pb";
import { getDeadline } from "../../../proto/deadline";
import { IStudentSubmission } from "../../models";

import { NavigationManager } from "../../managers/NavigationManager";

interface IPanelProps {
    course: Course;
    labs: IStudentSubmission[];
    navMan: NavigationManager;
}

// TODO(meling) why is there an 'any' generic type for React.Component? Is it needed?
class CoursePanel extends React.Component<IPanelProps, any> {

    public render() {
        const labPath: string = "app/student/courses/" + this.props.course.getId() + "/lab/";
        const glabPath: string = "app/student/courses/" + this.props.course.getId() + "/grouplab/";

        return (
            <div className="col-lg-3 col-md-6 col-sm-6">
                <div className="panel panel-primary">
                    <div className="panel-heading clickable"
                        onClick={() => this.handleCourseClick()}>{this.props.course.getName()}</div>
                    <div className="panel-body">
                        <DynamicTable
                            header={["Labs", "Score", "Deadline"]}
                            data={this.props.labs}
                            selector={(item: IStudentSubmission) => {
                                const score = item.latest ? (item.latest.score.toString() + "%") : "N/A";
                                return [
                                    item.assignment.getName(),
                                    score,
                                    getDeadline(item.assignment),
                                ];
                            }}
                            onRowClick={(lab: IStudentSubmission) => {
                                const path = !lab.assignment.getIsgrouplab() ? labPath : glabPath;
                                this.handleRowClick(path, lab.assignment);
                            }}
                        />
                    </div>
                </div>
            </div>
        );
    }

    private handleRowClick(pathPrefix: string, lab: Assignment) {
        if (lab) {
            this.props.navMan.navigateTo(pathPrefix + lab.getId());
        }
    }

    private handleCourseClick() {
        const uri: string = "app/student/courses/" + this.props.course.getId();
        this.props.navMan.navigateTo(uri);
    }
}

export { CoursePanel };
