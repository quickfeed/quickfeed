import * as React from "react";
import { IStudentLab } from "../../models";
import { LabResultView } from "../../pages/views/LabResultView";
import { User, Submission } from '../../../proto/ag_pb';

interface IStudentLabProps {
    studentSubmission: IStudentLab;
    student: User;
    courseURL: string;
    showApprove: boolean;
    slipdays: number;
    teacherPageView: boolean;
    courseCreatorView: boolean;
    reviewers: string[];
    onApproveClick: (approve: boolean) => void;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<boolean>;
}

export class StudentLab extends React.Component<IStudentLabProps> {
    public render() {
        return <LabResultView
            slipdays={this.props.slipdays}
            studentSubmission={this.props.studentSubmission}
            student={this.props.student}
            courseURL={this.props.courseURL}
            onApproveClick={this.props.onApproveClick}
            onRebuildClick={this.props.onRebuildClick}
            showApprove={this.props.showApprove}
            >
        </LabResultView>;
    }
}
