import * as React from "react";
import { Course, Enrollment } from "../../../proto/ag_pb";
import { DynamicTable } from "../../components";
import { IAssignmentLink } from "../../models";

export interface IEnrollmentViewProps {
    courses: IAssignmentLink[];
    onEnrollmentClick: (course: Course) => void;
}

export class EnrollmentView extends React.Component<IEnrollmentViewProps, {}> {
    public render() {
        return <DynamicTable
            data={this.props.courses}
            header={["Course code", "Course Name", "Action"]}
            selector={(course: IAssignmentLink) => this.createEnrollmentRow(course)}>
        </DynamicTable>;

    }

    public createEnrollmentRow(course: IAssignmentLink):
        Array<string | JSX.Element> {
        const base: Array<string | JSX.Element> = [course.course.getCode(), course.course.getName()];
        if (course.link) {
            if (course.link.getStatus() === Enrollment.UserStatus.STUDENT
                 || course.link.getStatus() === Enrollment.UserStatus.TEACHER) {
                base.push("");
            } else if (course.link.getStatus() === Enrollment.UserStatus.PENDING) {
                base.push("Pending");
            } else if (course.link.getStatus() === Enrollment.UserStatus.NONE) {
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
