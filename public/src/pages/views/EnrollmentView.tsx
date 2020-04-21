import * as React from "react";
import { Course, Enrollment } from "../../../proto/ag_pb";
import { DynamicTable } from "../../components";
import { IStudentLabsForCourse } from "../../models";

export interface IEnrollmentViewProps {
    courses: IStudentLabsForCourse[];
    onEnrollmentClick: (course: Course) => void;
}

export class EnrollmentView extends React.Component<IEnrollmentViewProps, {}> {
    public render() {
        return <DynamicTable
            data={this.props.courses}
            header={["Course code", "Course Name", "Status"]}
            selector={(course: IStudentLabsForCourse) => this.createEnrollmentRow(course)}>
        </DynamicTable>;

    }

    public createEnrollmentRow(course: IStudentLabsForCourse):
        (string | JSX.Element)[] {
        const base: (string | JSX.Element)[] = [course.course.getCode(), course.course.getName()];
        if (course.enrollment) {
            if (course.enrollment.getStatus() === Enrollment.UserStatus.STUDENT
                 || course.enrollment.getStatus() === Enrollment.UserStatus.TEACHER) {
                base.push("Enrolled");
            } else if (course.enrollment.getStatus() === Enrollment.UserStatus.PENDING) {
                base.push("Pending");
            } else if (course.enrollment.getStatus() === Enrollment.UserStatus.NONE) {
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
