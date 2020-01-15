import * as React from "react";
import { IStudentSubmission } from "../../models";
import { LabResultView } from "../../pages/views/LabResultView";

interface IStudentLabProps {
    assignment: IStudentSubmission;
    showApprove: boolean;
    onApproveClick: (approve: boolean) => void;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<boolean>;
}

export class StudentLab extends React.Component<IStudentLabProps> {
    public render() {
        return <LabResultView
            assignment={this.props.assignment}
            onApproveClick={this.props.onApproveClick}
            onRebuildClick={this.props.onRebuildClick}
            showApprove={this.props.showApprove}>
        </LabResultView>;
    }
}
