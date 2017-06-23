import * as React from "React";
import { IAssignment, ICourse } from "../../models";

interface IStudentLabProbs {
    course: ICourse;
    assignment: IAssignment;
}

class StudentLab extends React.Component<IStudentLabProbs, undefined> {
    public render() {
        return <h1>{this.props.assignment.name}</h1>;
    }
}

export { StudentLab, IStudentLabProbs };
