import * as React from "react";

import {DynamicTable} from "../../components";

import {IAssignment, ICourse} from "../../models";

import {NavigationManager} from "../../managers/NavigationManager";

interface IPanelProps {
    course: ICourse;
    labs: IAssignment[];
    navMan: NavigationManager;
}
class CoursePanel extends React.Component<IPanelProps, any> {

    public render() {
        const pathPrefix: string = "app/student/course/" + this.props.course.id + "/lab/";
        const rowLinks: { [lab: string]: string } = {};
        for (const lab of this.props.labs) {
            // lab.id will be used a key identifier for row_links map
            rowLinks[lab.id] = pathPrefix + lab.id;
        }

        return (
            <div className="col-lg-3 col-sm-6">
                <div className="panel panel-primary">
                    <div className="panel-heading clickable"
                         onClick={() => this.handleCourseClick()}>{this.props.course.name}</div>
                    <div className="panel-body">
                        <DynamicTable
                            header={["Labs", "Score", "Weight"]}
                            data={this.props.labs}
                            selector={(item: IAssignment) => [item.name, "50%", "100%"]}
                            onRowClick={(row) => this.handleRowClick(row)}
                            row_links={rowLinks}
                            link_key_identifier="id"
                        />
                    </div>
                </div>
            </div>
        );
    }

    private handleRowClick(path: string) {
        if (path) {
            this.props.navMan.navigateTo(path);
        }
    }

    private handleCourseClick() {
        const uri: string = "app/student/course/" + this.props.course.id;
        this.props.navMan.navigateTo(uri);
    }
}

export {CoursePanel};
