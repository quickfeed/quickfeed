import * as React from "react";
import { CoursePanel, Row } from "../../components";
import { NavigationManager } from "../../managers/NavigationManager";
import { IStudentLabsForCourse } from "../../models";

interface ICourseOverviewProps {
    courseOverview: IStudentLabsForCourse[];
    groupCourseOverview: IStudentLabsForCourse[];
    navMan: NavigationManager;
}

export class CoursesOverview extends React.Component<ICourseOverviewProps> {

    public render() {
        const groupCourses = this.props.groupCourseOverview;
        const courses = this.props.courseOverview.map((val, key) => {
            if (groupCourses && groupCourses[key] && groupCourses[key].course.getId() === val.course.getId()) {
                for (let iter = 0; iter < val.labs.length; iter++) {
                    if (val.labs[iter].assignment.getIsgrouplab()) {
                        val.labs[iter].submission = groupCourses[key].labs[iter].submission;
                    }
                }
            }
            return <CoursePanel
                key={key}
                course={val.course}
                labs={val.labs}
                navMan={this.props.navMan} />;
        });

        // TODO(meling) WTF does this code do?
        let added = 0;
        let l = courses.length;
        for (let index = 1; index < l; index++) {
            if (index % 4 === 0) {
                courses.splice(index + added, 0,
                    <div
                        key={index * 10000}
                        className="visible-md-block visible-sm-block visible-lg-block clearfix">
                    </div>,
                );
                l += 1;
                added += 1;
            } else if (index % 2 === 0) {
                courses.splice(index + added, 0,
                    <div
                        key={index * 10000}
                        className="visible-md-block visible-sm-block clearfix">
                    </div>);
                l += 1;
                added += 1;
            }
        }

        return <div>
            <h1>Your Courses</h1>
            <Row>{courses}</Row>
        </div>;
    }
}
