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
    onApproveClick: (approve: boolean) => void;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<boolean>;
    setApproved?: (submissionID: number, status: Submission.Status) => void;
    setReady?: (submissionID: number, ready: boolean) => void;
    getReviewers: (submissionID: number) => Promise<string[]>;
}

export class StudentLab extends React.Component<IStudentLabProps> {
    public render() {
        return <LabResultView
            slipdays={this.props.slipdays}
            studentSubmission={this.props.studentSubmission}
            student={this.props.student}
            courseURL={this.props.courseURL}
            teacherPageView={this.props.teacherPageView}
            courseCreatorView={this.props.courseCreatorView}
            onApproveClick={this.props.onApproveClick}
            onRebuildClick={this.props.onRebuildClick}
            showApprove={this.props.showApprove}
            setApproved={this.props.setApproved}
            setReady={this.props.setReady}
            getReviewers={this.props.getReviewers}
            >
        </LabResultView>;
    }
}
