import * as React from "react";
// import { Table } from "react-bootstrap";
import { Row } from "../../components";

interface ILastBuildInfo {
    submission_id: number;
    pass_tests: number;
    fail_tests: number;
    exec_time: number;
    build_time: string;
    build_id: number;
    isApproved: boolean;
    showApprove: boolean;
    onApproveClick: () => void;
    onRebuildClick: (submissionID: number) => Promise<boolean>;
}

interface ILastBuildInfoState {
    rebuilding: boolean;
    approved: boolean;
}

export class LastBuildInfo extends React.Component<ILastBuildInfo, ILastBuildInfoState> {

    constructor(props: ILastBuildInfo) {
        super(props);
        this.state = {
            rebuilding: false,
            approved: this.props.isApproved,
         };
    }
    public render() {
        let approveButton = <p></p>;
        if (this.props.showApprove) {
            approveButton = <p> <button type="button"
                id="approve"
                className={this.setButtonColor("approve")}
                onClick={this.state.approved ?
                    () => { console.log("Already approved"); } : () => this.approve()}>
                     {this.setButtonString("approve")} </button> </p>; }

        return (
            <Row>
                <div className="col-lg-8">
                    <h2>Latest build</h2>
                    <p id="passes">Passed tests:  {this.props.pass_tests}</p>
                    <p id="fails">Failed tests:  {this.props.fail_tests}</p>
                    <p id="buildtime">Execution time:  {this.props.exec_time / 1000} s</p>
                    <p id="timedate">Build date:  {this.props.build_time ? this.props.build_time.toString() : "-"}</p>
                    <p id="buildid">Build ID: {this.props.build_id}</p>
                </div>
                <div className="col-lg-4 hidden-print">
                    <h2>Actions</h2>
                    <Row>
                        <div className="col-lg-12">
                            <p>
                                <button type="button" id="rebuild" className={this.setButtonColor("rebuild")}
                                    onClick={
                                        this.state.rebuilding ? () => {console.log("Rebuilding..."); }
                                         : () => {this.rebuildSubmission(); }
                                    }>{this.setButtonString("rebuild")}</button>
                            </p>
                            {approveButton}
                        </div>
                    </Row>
                </div>
            </Row>
        );
    }

    private async rebuildSubmission() {
        console.log("Rebuilds are disabled");
        /*this.setState({
            rebuilding: true,
        });
        await this.props.onRebuildClick(this.props.submission_id).then(() => {
            this.setState({
                rebuilding: false,
            });
        });*/
    }

    private async approve() {
        await this.props.onApproveClick();
        this.setState({
            approved: true,
        });
    }

    private setButtonColor(id: string): string {
        switch (id) {
            case "rebuild" : {
                return this.state.rebuilding ? "btn btn-secondary" : "btn btn-primary";
            }
            case "approve" : {
                return this.state.approved ? "btn btn-success" : "btn btn-primary";
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
                return this.state.approved ? "Approved" : "Approve";
            }
            default : {
                return "";
            }
        }
    }
}
