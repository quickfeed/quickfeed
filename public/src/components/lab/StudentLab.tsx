import * as React from "react";
import { Course, User } from "../../../proto/ag_pb";
import { IStudentSubmission } from "../../models";
import { LabResultView } from "../../pages/views/LabResultView";

interface IStudentLabProbs {
    course: Course;
    assignment: IStudentSubmission;
    student?: User;
    showApprove: boolean;
    onApproveClick: () => void;
    onRebuildClick: () => void;
}

export class StudentLab extends React.Component<IStudentLabProbs> {
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
