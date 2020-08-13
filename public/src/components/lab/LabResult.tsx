import * as React from "react";
import { ProgressBar, Row } from "../../components";
import { Submission } from "../../../proto/ag_pb";

interface ILabResultProps {
    assignment_id: number;
    submission_id: number;
    progress: number;
    status: Submission.Status;
    lab: string;
    comment: string;
    authorName?: string;
    teacherView: boolean;
    onSubmissionUpdate: (status: Submission.Status, comment: string) => void;
    onSubmissionRebuild: (assignmentID: number, submissionID: number) => Promise<boolean>;
}

interface ILabResultState {
    rebuilding: boolean;
    commenting: boolean;
    comment: string;
}

export class LabResult extends React.Component<ILabResultProps, ILabResultState> {

    constructor(props: ILabResultProps) {
        super(props);
        this.state = {
            rebuilding: false,
            commenting: false,
            comment: props.comment,
        };
    }

    public render() {
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
                    {this.props.teacherView ? this.actionButtons() : null}
                    {this.props.teacherView ? this.commentDiv() : null}
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

    public commentDiv(): JSX.Element {
        const editComment = <div className="row lab-comment input-group">
            <input
                className="form-control lab-input"
                autoFocus={true}
                type="text"
                defaultValue={this.props.comment}
                onChange={(e) => this.setNewComment(e.target.value)}
                onBlur={() => this.toggleCommenting()}
                onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                        this.props.onSubmissionUpdate(this.props.status, this.state.comment);
                        this.setState({
                            commenting: false,
                        });
                    } else if (e.key === 'Escape') {
                        this.toggleCommenting();
                    }
                }}
            />
            {this.state.comment}</div>
        const showComment = <div className="row lab-comment"
            onClick={() => this.toggleCommenting()}
        >{this.props.comment}</div>;
        return this.state.commenting ? editComment : showComment;
    }

    public actionButtons(): JSX.Element {
        const approveButton = <button type="button" className={this.setButtonClassColor("approve")}
            onClick={
                () => {this.props.onSubmissionUpdate(Submission.Status.APPROVED, this.props.comment); }
            }
        >{this.setButtonString("approve")}</button>;
        const revisionButton = <button type="button" className={this.setButtonClassColor("revision")}
            onClick={
                () => {this.props.onSubmissionUpdate(Submission.Status.REVISION, this.props.comment); }
            }
        >{this.setButtonString("revision")}</button>;
        const rejectButton = <button type="button" className={this.setButtonClassColor("reject")}
            onClick={
                () => {this.props.onSubmissionUpdate(Submission.Status.REJECTED, this.props.comment); }
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

    private toggleCommenting() {
        this.setState((prevState: ILabResultState) => ({
            commenting: !prevState.commenting,
        }));
    }

    private setNewComment(input: string) {
        this.setState({
            comment: input,
        });
    }
}
