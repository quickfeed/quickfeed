import * as React from "react";
import { DynamicTable } from "../../components";
import { CourseUserState, ICourse, ICourseUser, IStudentCourse, IUser } from "../../models";

import { ArrayHelper } from "../../helper";

interface IEnrollmentViewProps {
    courses: IStudentCourse[];
    onEnrollmentClick: (course: ICourse) => void;
}

class EnrollmentView extends React.Component<IEnrollmentViewProps, {}> {
    public render() {
        return <DynamicTable
            data={this.props.courses}
            header={["Course tag", "Course Name", "Action"]}
            selector={(course: IStudentCourse) => this.createEnrollmentRow(this.props.courses, course)}>
        </DynamicTable>;

    }

    public createEnrollmentRow(studentCourses: IStudentCourse[], course: IStudentCourse): Array<string | JSX.Element> {
        const base: Array<string | JSX.Element> = [course.course.tag, course.course.name];
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

export { EnrollmentView, IEnrollmentViewProps };
