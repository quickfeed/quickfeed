import * as React from "react";
import { DynamicTable } from "../../components";
import { CourseStudentState, ICourse, ICourseStudent, IUser } from "../../models";

import { ArrayHelper } from "../../helper";

interface IEnrollmentViewProps {
    courses: ICourse[];
    studentCourses: ICourseStudent[];
    curUser: IUser | null;
    onEnrollmentClick: (user: IUser, course: ICourse) => void;
}

class EnrollmentView extends React.Component<IEnrollmentViewProps, {}> {
    public render() {
        return <DynamicTable
            data={this.props.courses}
            header={["Course tag", "Course Name", "Action"]}
            selector={(course: ICourse) => this.createEnrollmentRow(this.props.studentCourses, course)}>
        </DynamicTable>;

    }

    public createEnrollmentRow(studentCourses: ICourseStudent[], course: ICourse): Array<string | JSX.Element> {
        const base: Array<string | JSX.Element> = [course.tag, course.name];
        const curUser = this.props.curUser;
        if (!curUser) {
            return base;
        }
        const temp = ArrayHelper.find(studentCourses, (a: ICourseStudent) => a.courseId === course.id);
        if (temp) {
            if (temp.state === CourseStudentState.accepted) {
                base.push("Enrolled");
            } else if (temp.state === CourseStudentState.pending) {
                base.push("Pending");
            } else {
                base.push(<div>
                    <button
                        onClick={() => { this.props.onEnrollmentClick(curUser, course); }}
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
                    onClick={() => { this.props.onEnrollmentClick(curUser, course); }}
                    className="btn btn-primary">
                    Enroll
                </button>);
        }
        return base;
    }
}

export { EnrollmentView, IEnrollmentViewProps };
