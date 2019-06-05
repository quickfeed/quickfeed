import * as React from "react";
import { DynamicTable } from "../../components";
import { ICourse, ICourseWithEnrollStatus, IUserCourse } from "../../models";
import { Enrollment } from "../../../proto/ag_pb";

export interface IEnrollmentViewProps {
    courses: IUserCourse[];
    onEnrollmentClick: (course: ICourse) => void;
}

export class EnrollmentView extends React.Component<IEnrollmentViewProps, {}> {
    public render() {
        return <DynamicTable
            data={this.props.courses}
            header={["Course code", "Course Name", "Action"]}
            selector={(course: IUserCourse) => this.createEnrollmentRow(this.props.courses, course)}>
        </DynamicTable>;

    }

    public createEnrollmentRow(studentCourses: IUserCourse[], course: IUserCourse): Array<string | JSX.Element> {
        const base: Array<string | JSX.Element> = [course.course.code, course.course.name];
        if (course.link) {
            if (course.link.state === Enrollment.UserStatus.STUDENT || course.link.state === Enrollment.UserStatus.TEACHER) {
                base.push("Enrolled");
            } else if (course.link.state === Enrollment.UserStatus.PENDING) {
                base.push("Pending");
            } else if (course.link.state === Enrollment.UserStatus.NONE) {
                base.push(
                    <button
                        onClick={() => { this.props.onEnrollmentClick(course.course); }}
                        className="btn btn-primary">
                        Enroll
                    </button>);
            } else {
                base.push(
                    <span style={{ padding: "7px", verticalAlign: "middle" }} className="bg-danger">
                        Rejected
                    </span>);
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
