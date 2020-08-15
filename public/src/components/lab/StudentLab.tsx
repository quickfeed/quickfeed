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
    commenting: boolean;
    updateSubmissionStatus: (status: Submission.Status) => void;
    setSubmissionComment: (comment: string) => void;
    onSubmissionRebuild: (assignmentID: number, submissionID: number) => Promise<boolean>;
    toggleCommenting: (toggleOn: boolean) => void;
}

export class StudentLab extends React.Component<IStudentLabProps> {
    public render() {
        return <LabResultView
            slipdays={this.props.slipdays}
            submissionLink={this.props.studentSubmission}
            student={this.props.student}
            courseURL={this.props.courseURL}
            commenting={this.props.commenting}
            updateSubmissionStatus={this.props.updateSubmissionStatus}
            setSubmissionComment={this.props.setSubmissionComment}
            rebuildSubmission={this.props.onSubmissionRebuild}
            toggleCommenting={this.props.toggleCommenting}
            teacherPageView={this.props.teacherPageView}
            >
        </LabResultView>;
    }
}
