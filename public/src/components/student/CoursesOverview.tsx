import * as React from "react";

import {CoursePanel, Row} from "../../components";

import {ICoursesWithAssignments} from "../../models";

import {NavigationManager} from "../../managers/NavigationManager";

interface ICourseOverviewProps {
    course_overview: ICoursesWithAssignments[];
    navMan: NavigationManager;
}

class CoursesOverview extends React.Component<ICourseOverviewProps, any> {

    public render() {
        const courses = this.props.course_overview.map((val, key) => {
            return <CoursePanel course={val.course} labs={val.labs} navMan={this.props.navMan}/>;
        });

        let added: number = 0;
        let index: number = 1;
        let l: number = courses.length;
        for (index; index < l; index++) {
            if (index % 2 === 0) {
                courses.splice(index + added, 0, <div className="visible-md-block visible-sm-block clearfix"></div>);
                l += 1;
                added += 1;
            }
            if (index % 4 === 0) {
                courses.splice(index + added, 0, <div className="visible-lg-block clearfix"></div>);
                l += 1;
                added += 1;
            }
        }

        return (
            <div>
                <h1>Your Courses</h1>
                <Row>{courses}</Row>
            </div>
        );
    }
}

export {CoursesOverview};
