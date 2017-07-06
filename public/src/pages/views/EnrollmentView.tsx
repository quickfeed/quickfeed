import * as React from "react";
import { DynamicTable } from "../../components";
import { CourseUserState, ICourse, ICourseUserLink, IUser, IUserCourse } from "../../models";

import { ArrayHelper } from "../../helper";

export interface IEnrollmentViewProps {
    courses: IUserCourse[];
    onEnrollmentClick: (course: ICourse) => void;
}

export class EnrollmentView extends React.Component<IEnrollmentViewProps, {}> {
    public render() {
        return <DynamicTable
            data={this.props.courses}
            header={["Course tag", "Course Name", "Action"]}
            selector={(course: IUserCourse) => this.createEnrollmentRow(this.props.courses, course)}>
        </DynamicTable>;

    }

    public createEnrollmentRow(studentCourses: IUserCourse[], course: IUserCourse): Array<string | JSX.Element> {
        const base: Array<string | JSX.Element> = [course.course.code, course.course.name];
        if (course.link) {
            if (course.link.state === CourseUserState.student) {
                base.push("Enrolled");
            } else if (course.link.state === CourseUserState.pending) {
                base.push("Pending");
            } else {
                base.push(<div>
                    <button
                        onClick={() => { this.props.onEnrollmentClick(course.course); }}
                        className="btn btn-primary">
                        Enroll
                    </button>
                    <span style={{ padding: "7px", verticalAlign: "middle" }} className="bg-danger">
                        Rejected
                    </span>
                </div>);
            }
        } else {
            base.push(
                <button
                    onClick={() => { this.props.onEnrollmentClick(course.course); }}
                    className="btn btn-primary">
                    Enroll
                </button>);
        }
        return base;
    }
}
