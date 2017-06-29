import * as React from "react";

import { CoursePanel, Row } from "../../components";

import { ICoursesWithAssignments } from "../../models";

import { NavigationManager } from "../../managers/NavigationManager";

interface ICourseOverviewProps {
    course_overview: ICoursesWithAssignments[];
    navMan: NavigationManager;
}

class CoursesOverview extends React.Component<ICourseOverviewProps, any> {

    public render() {
        const courses = this.props.course_overview.map((val, key) => {
            return <CoursePanel course={val.course} labs={val.labs} navMan={this.props.navMan} />;
        });
        const count = 4;
        let index: number = count;
        let l: number = courses.length;
        for (index; index < l; index += count) {
            console.log("index", index);
            courses.splice(index, 0, <div className="visible-lg-block visible-md-block clearfix"></div>);
            l += 1;
            index += 1;
        }

        return (
            <div>
                <h1>Your Courses</h1>
                <Row>{courses}</Row>
            </div>
        );
    }
}

export { CoursesOverview };
