import * as React from "react";
import { IStudentLab } from "../../models";
import { LabResultView } from "../../pages/views/LabResultView";

interface IStudentLabProps {
    assignment: IStudentLab;
    showApprove: boolean;
    slipdays: number;
    onApproveClick: (approve: boolean) => void;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<boolean>;
    getReviewers: (assignmentID: number, submissionID: number) => string[];
}

export class StudentLab extends React.Component<IStudentLabProps> {
    public render() {
        return <LabResultView
            slipdays={this.props.slipdays}
            studentSubmission={this.props.assignment}
            onApproveClick={this.props.onApproveClick}
            onRebuildClick={this.props.onRebuildClick}
            showApprove={this.props.showApprove}>
        </LabResultView>;
    }
}
