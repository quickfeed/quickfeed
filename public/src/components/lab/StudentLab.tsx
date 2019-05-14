import * as React from "react";
import { IAssignment, ICourse, IStudentSubmission, ISubmission, ITestCases, IUser } from "../../models";
import { LabResultView } from "../../pages/views/LabResultView";
import { User } from "../../../proto/ag_pb";

interface IStudentLabProbs {
    course: ICourse;
    assignment: IStudentSubmission;
    student?: User;
    showApprove: boolean;
    onApproveClick: () => void;
    onRebuildClick: () => void;
}

class StudentLab extends React.Component<IStudentLabProbs, {}> {
    public render() {
        return <LabResultView
            course={this.props.course}
            labInfo={this.props.assignment}
            onApproveClick={this.props.onApproveClick}
            onRebuildClick={this.props.onRebuildClick}
            showApprove={this.props.showApprove}>
        </LabResultView>;
    }
}

export { StudentLab, IStudentLabProbs };
