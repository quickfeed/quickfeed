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
    onSubmissionStatusUpdate: (status: Submission.Status) => void;
    onSubmissionRebuild: (assignmentID: number, submissionID: number) => Promise<boolean>;
}

export class StudentLab extends React.Component<IStudentLabProps> {
    public render() {
        return <LabResultView
            slipdays={this.props.slipdays}
            submissionLink={this.props.studentSubmission}
            student={this.props.student}
            courseURL={this.props.courseURL}
            onSubmissionStatusUpdate={this.props.onSubmissionStatusUpdate}
            onSubmissionRebuild={this.props.onSubmissionRebuild}
            teacherPageView={this.props.teacherPageView}
            >
        </LabResultView>;
    }
}
