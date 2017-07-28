import * as React from "react";
import { IAssignment, ICourse, IStudentSubmission, ISubmission, ITestCases, IUser } from "../../models";
import { LabResultView } from "../../pages/views/LabResultView";

interface IStudentLabProbs {
    course: ICourse;
    assignment: IStudentSubmission;
    student?: IUser;
}

class StudentLab extends React.Component<IStudentLabProbs, {}> {
    public render() {
        return <LabResultView course={this.props.course} labInfo={this.props.assignment}></LabResultView>;
    }
}

export { StudentLab, IStudentLabProbs };
