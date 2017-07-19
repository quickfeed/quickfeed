import * as React from "react";
import { DynamicTable } from "../../components";
import { CourseUserState, ICourse, ICourseWithEnrollStatus } from "../../models";

export interface IEnrollmentViewProps {
    courses: ICourseWithEnrollStatus[];
    onEnrollmentClick: (course: ICourse) => void;
}

export class EnrollmentView extends React.Component<IEnrollmentViewProps, {}> {
    public render() {
        return <DynamicTable
            data={this.props.courses}
            header={["Course tag", "Course Name", "Action"]}
            selector={(course: ICourseWithEnrollStatus) => this.createEnrollmentRow(this.props.courses, course)}>
        </DynamicTable>;

    }

    public createEnrollmentRow(studentCourses: ICourseWithEnrollStatus[],
                               course: ICourseWithEnrollStatus): Array<string | JSX.Element> {
        const base: Array<string | JSX.Element> = [course.code, course.name];
        if (course.enrolled >= 0) {
            if (course.enrolled === CourseUserState.student) {
                base.push("Enrolled");
            } else if (course.enrolled === CourseUserState.pending) {
                base.push("Pending");
            } else {
                base.push(<div>
                    <button
                        onClick={() => {
                            this.props.onEnrollmentClick(course);
                        }}
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
                    onClick={() => {
                        this.props.onEnrollmentClick(course);
                    }}
                    className="btn btn-primary">
                    Enroll
                </button>);
        }
        return base;
    }
}
