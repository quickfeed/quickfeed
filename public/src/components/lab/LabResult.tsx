import * as React from "react";
import { ProgressBar, Row } from "../../components";

interface ILabResult {
    assignment_id: number;
    submission_id: number;
    progress: number;
    lab: string;
    authorName?: string;
    showApprove: boolean;
    isApproved: boolean;
    onApproveClick: (approve: boolean) => void;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<boolean>;
}

interface ILabResultState {
    rebuilding: boolean;
}

export class LabResult extends React.Component<ILabResult, ILabResultState> {

    constructor(props: ILabResult) {
        super(props);
        this.state = {
            rebuilding: false,
        };
    }

    public render() {
        let approveButton = <div></div>;
        let rebuildButton = <div></div>;
        if (this.props.showApprove) {
            approveButton = <div className="btn lab-btn approve-btn">
                <button type="button"
                id="approve"
                className={this.setButtonColor("approve")}
                title={this.setTooltip()}
                onClick={() => this.approve()}>
                     {this.setButtonString("approve")}
                </button></div>;
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

    private async approve() {
        this.props.onApproveClick(!this.props.isApproved);
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
