import * as React from "react";

import { CoursePanel, Row } from "../../components";

import { IAssignmentLink, IStudentSubmission } from "../../models";

import { NavigationManager } from "../../managers/NavigationManager";

interface ICourseOverviewProps {
    courseOverview: IAssignmentLink[];
    groupCourseOverview: IAssignmentLink[];
    navMan: NavigationManager;
}

class CoursesOverview extends React.Component<ICourseOverviewProps, any> {

    public render() {
        const groupCourses = this.props.groupCourseOverview ? this.props.groupCourseOverview : null;
        const courses = this.props.courseOverview.map((val, key) => {
            const courseAssignments: IStudentSubmission[] = val.assignments;
            if (groupCourses && groupCourses[key] && groupCourses[key].course.getId() === val.course.getId()) {

                for (let iter = 0; iter < courseAssignments.length; iter++) {
                    if (courseAssignments[iter].assignment.getIsgrouplab()) {
                        courseAssignments[iter].latest = groupCourses[key].assignments[iter].latest;
                    }
                }
            }
            return <CoursePanel
                key={key}
                course={val.course}
                labs={courseAssignments}
                navMan={this.props.navMan} />;
        });

        let added: number = 0;
        let index: number = 1;
        let l: number = courses.length;
        for (index; index < l; index++) {
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

        return (
            <div>
                <h1>Your Courses</h1>
                <Row>{courses}</Row>
            </div>
        );
    }
}

export { CoursesOverview };
