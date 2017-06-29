import * as React from "react";
import { DynamicTable } from "../../components";
import { ICourse, IUser } from "../../models";

import { ArrayHelper } from "../../helper";

interface IEnrollmentViewProps {
    courses: ICourse[];
    studentCourses: ICourse[];
    curUser: IUser | null;
    onEnrollmentClick: (user: IUser, course: ICourse) => void;
}

class EnrollmentView extends React.Component<IEnrollmentViewProps, undefined> {
    public render() {
        return <DynamicTable
            data={this.props.courses}
            header={["Course tag", "Course Name", "Action"]}
            selector={(course: ICourse) => this.createEnrollmentRow(this.props.studentCourses, course)}>
        </DynamicTable>;

    }

    public createEnrollmentRow(studentCourses: ICourse[], course: ICourse): Array<string | JSX.Element> {
        const base: Array<string | JSX.Element> = [course.tag, course.name];
        const curUser = this.props.curUser;
        if (!curUser) {
            return base;
        }
        if (!ArrayHelper.find(studentCourses, (a: ICourse) => a.id === course.id)) {
            base.push(<button
                onClick={() => { this.props.onEnrollmentClick(curUser, course); }}
                className="btn btn-primary">
                Enroll
                    </button>);
        } else {
            base.push("Enrolled");
        }
        return base;
    }
}

export { EnrollmentView, IEnrollmentViewProps };
