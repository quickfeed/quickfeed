import * as React from "react";
import { ProgressBar, Row } from "../../components";
import { Submission } from "../../../proto/ag_pb";

interface ILabResultProps {
    assignment_id: number;
    submission_id: number;
    progress: number;
    status: Submission.Status;
    lab: string;
    authorName?: string;
    teacherView: boolean;
    onSubmissionStatusUpdate: (status: Submission.Status) => void;
    onSubmissionRebuild: (assignmentID: number, submissionID: number) => Promise<boolean>;
}

interface ILabResultState {
    rebuilding: boolean;
}

export class LabResult extends React.Component<ILabResultProps, ILabResultState> {

    constructor(props: ILabResultProps) {
        super(props);
        this.state = {
            rebuilding: false,
        };
    }

    public render() {
        let buttonDiv = <div></div>;
        if (this.props.teacherView) {
            buttonDiv = this.actionButtons();
        }

        let labHeading: JSX.Element;
        if (this.props.authorName) {
            labHeading = <h3>{this.props.authorName + ": "} {this.props.lab}</h3>;
        } else {
            labHeading = <div>
                <p className="lead">Your progress on <strong><span
                    id="lab-headline">{this.props.lab}</span></strong>
                </p>
            </div>;
        }
        return (
                <div className="col-lg-12">
                    <Row>
                    {labHeading}
                    <ProgressBar progress={this.props.progress}></ProgressBar></Row>
                    <Row>{buttonDiv}</Row>
            </div>
        );
    }

    private async rebuildSubmission() {
        this.setState({
            rebuilding: true,
        });
        await this.props.onSubmissionRebuild(this.props.assignment_id, this.props.submission_id).then(() => {
            this.setState({
                rebuilding: false,
            });
        });
    }

    public actionButtons(): JSX.Element {
        const approveButton = <button type="button" className={this.setButtonClassColor("approve")}
            onClick={
                () => {this.props.onSubmissionStatusUpdate(Submission.Status.APPROVED); }
            }
        >{this.setButtonString("approve")}</button>;
        const revisionButton = <button type="button" className={this.setButtonClassColor("revision")}
            onClick={
                () => {this.props.onSubmissionStatusUpdate(Submission.Status.REVISION); }
            }
        >{this.setButtonString("revision")}</button>;
        const rejectButton = <button type="button" className={this.setButtonClassColor("reject")}
            onClick={
                () => {this.props.onSubmissionStatusUpdate(Submission.Status.REJECTED); }
            }
        >{this.setButtonString("reject")}</button>;
        const rebuildButton = <button type="button" className={this.setButtonClassColor("rebuild")}
            onClick={
                this.state.rebuilding ? () => {console.log("Rebuilding..."); } : () => {this.rebuildSubmission(); }
            }
        >{this.setButtonString("rebuild")}</button>;

        return <div className="row lab-btns">{approveButton}{revisionButton}{rejectButton}{rebuildButton}</div>;
    }

    private setButtonClassColor(id: string): string {
        switch (id) {
            case "rebuild" : {
                return this.state.rebuilding ? "btn lab-btn btn-secondary" : "btn lab-btn btn-primary";
            }
            case "approve" : {
                return this.props.status === Submission.Status.APPROVED ? "btn lab-btn btn-success" : "btn lab-btn btn-default";
            }
            case "reject" : {
                return this.props.status === Submission.Status.REJECTED ? "btn lab-btn btn-danger" : "btn lab-btn btn-default";
            }
            case "revision" : {
                return this.props.status === Submission.Status.REVISION ? "btn lab-btn btn-warning" : "btn lab-btn btn-default"
            }
            default: {
                return "";
            }
        }
    }

    private setButtonString(id: string): string {
        switch (id) {
            case "rebuild" : {
                return this.state.rebuilding ? "Rebuilding" : "Rebuild";
            }
            case "approve" : {
                return this.props.status === Submission.Status.APPROVED ? "Approved" : "Approve";
            }
            case "reject" : {
                return this.props.status === Submission.Status.REJECTED ? "Failed" : "Fail";
            }
            case "revision" : {
                return this.props.status === Submission.Status.REVISION ? "Revising" : "Revision";
            }
            default : {
                return "";
            }
        }
    }

}
