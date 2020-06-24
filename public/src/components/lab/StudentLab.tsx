import * as React from "react";
import { ISubmissionLink } from "../../models";
import { LabResultView } from "../../pages/views/LabResultView";
import { User, Submission } from "../../../proto/ag_pb";

interface IStudentLabProps {
    studentSubmission: ISubmissionLink;
    student: User;
    courseURL: string;
    slipdays: number;
    teacherPageView: boolean;
    onApproveClick: (status: Submission.Status, approve: boolean) => Promise<boolean>;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<boolean>;
}

export class StudentLab extends React.Component<IStudentLabProps> {
    public render() {
        return <LabResultView
            slipdays={this.props.slipdays}
            submissionLink={this.props.studentSubmission}
            student={this.props.student}
            courseURL={this.props.courseURL}
            onApproveClick={this.props.onApproveClick}
            onRebuildClick={this.props.onRebuildClick}
            teacherPageView={this.props.teacherPageView}
            >
        </LabResultView>;
    }
}
