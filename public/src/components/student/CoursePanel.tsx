import * as React from "react";

import { DynamicTable } from "../../components";

import { IAssignment, ICourse, IStudentSubmission } from "../../models";

import { NavigationManager } from "../../managers/NavigationManager";

interface IPanelProps {
    course: ICourse;
    labs: IStudentSubmission[];
    navMan: NavigationManager;
}
class CoursePanel extends React.Component<IPanelProps, any> {

    public render() {
        const pathPrefix: string = "app/student/courses/" + this.props.course.id + "/lab/";

        return (
            <div className="col-lg-3 col-md-6 col-sm-6">
                <div className="panel panel-primary">
                    <div className="panel-heading clickable"
                        onClick={() => this.handleCourseClick()}>{this.props.course.name}</div>
                    <div className="panel-body">
                        <DynamicTable
                            header={["Labs", "Score", "Deadline"]}
                            data={this.props.labs}
                            selector={(item: IStudentSubmission) => {
                                const score = item.latest ? (item.latest.score.toString() + "%") : "N/A";
                                return [
                                    item.assignment.name,
                                    score,
                                    item.assignment.deadline.toDateString(),
                                ];
                            }}
                            onRowClick={(lab: IStudentSubmission) => this.handleRowClick(pathPrefix, lab.assignment)}
                        />
                    </div>
                </div>
            </div>
        );
    }

    private handleRowClick(pathPrefix: string, lab: IAssignment) {
        if (lab) {
            this.props.navMan.navigateTo(pathPrefix + lab.id);
        }
    }

    private handleCourseClick() {
        const uri: string = "app/student/course/" + this.props.course.id;
        this.props.navMan.navigateTo(uri);
    }
}

export { CoursePanel };
