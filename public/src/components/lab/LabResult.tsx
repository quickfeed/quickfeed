import * as React from "react";
import { ProgressBar, Row } from "../../components";
import { Submission } from "../../../proto/ag_pb";
import { submissionStatusSelector } from "../../componentHelper";

interface ILabResult {
    assignment_id: number;
    submission_id: number;
    progress: number;
    status: Submission.Status;
    lab: string;
    authorName?: string;
    teacherView: boolean;
    isApproved: boolean;
    onApproveClick: (status: Submission.Status, approve: boolean) => Promise<boolean>;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<boolean>;
}

interface ILabResultState {
    rebuilding: boolean;
    status: Submission.Status;
}

export class LabResult extends React.Component<ILabResult, ILabResultState> {

    constructor(props: ILabResult) {
        super(props);
        this.state = {
            rebuilding: false,
            status: this.props.status,
        };
    }

    public render() {
        let approveButton = <div></div>;
        let rebuildButton = <div></div>;
        if (this.props.teacherView) {
            approveButton = submissionStatusSelector(this.props.status, (action: string) => this.approve(action), "approve-btn")
            rebuildButton = <div className="btn lab-btn rebuild-btn">
            <button type="button" id="rebuild" className={this.setButtonColor("rebuild")}
                onClick={
                    this.state.rebuilding ? () => {console.log("Rebuilding..."); }
                     : () => {this.rebuildSubmission(); }
                }>{this.setButtonString("rebuild")}
                </button>
        </div>;
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
                    <Row>{approveButton} {rebuildButton}</Row>
            </div>
        );
    }

    private async rebuildSubmission() {
        this.setState({
            rebuilding: true,
        });
        await this.props.onRebuildClick(this.props.assignment_id, this.props.submission_id).then(() => {
            this.setState({
                rebuilding: false,
            });
        });
    }

    private async approve(action: string) {
        let newStatus: Submission.Status = Submission.Status.NONE;
        let newBool = false;
        switch (action) {
            case "1":
                newStatus = Submission.Status.APPROVED;
                newBool = true;
                break;
            case "2":
                newStatus = Submission.Status.REJECTED;
                break;
            case "3":
                newStatus = Submission.Status.REVISION;
                break;
            default:
                newStatus = Submission.Status.NONE;
                break;
        }
        const ans = await this.props.onApproveClick(newStatus, newBool);
        if (ans) {
            this.setState({
                status: newStatus,
            });
        }
    }

    private setButtonColor(id: string): string {
        switch (id) {
            case "rebuild" : {
                return this.state.rebuilding ? "btn btn-secondary" : "btn btn-primary";
            }
            case "approve" : {
                return this.props.isApproved ? "btn btn-success" : "btn btn-primary";
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
                return this.props.isApproved ? "Approved" : "Approve";
            }
            default : {
                return "";
            }
        }
    }

    private setTooltip(): string {
        return this.props.isApproved ? "Undo approval" : "Approve";
    }
}
