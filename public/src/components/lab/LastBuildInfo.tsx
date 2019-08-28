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
    showApprove: boolean;
    onApproveClick: () => void;
    onRebuildClick: (submissionID: number) => void;
}

interface ILastBuildInfoState {
    rebuilding: boolean;
}

export class LastBuildInfo extends React.Component<ILastBuildInfo, ILastBuildInfoState> {

    public render() {
        this.state = { rebuilding: false };
        let approveButton = <p></p>;
        if (this.props.showApprove) {
            approveButton = <p> <button type="button"
                id="approve"
                className="btn btn-primary"
                onClick={() => this.handleClick(this.props.onApproveClick)}> Approve </button> </p>;
        }
        // const table = <Table striped borderless size="sm">
        //     <tbody>
        //         <tr><td>Tests passed</td><td id="passes">{this.props.pass_tests}</td></tr>
        //         <tr><td>Tests failed</td><td id="fails">{this.props.fail_tests}</td></tr>
        //         <tr><td>Execution time</td><td id="buildtime">{this.props.exec_time / 1000} s</td></tr>
        //         <tr><td>Build date</td><td id="timedate">{this.props.build_time.toString()}</td></tr>
        //         <tr><td>Build ID</td><td id="buildid">{this.props.build_id}</td></tr>
        //     </tbody>
        // </Table>;
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
                                <button type="button" id="rebuild" className="btn btn-primary"
                                    onClick={() => {
                                        this.setState({
                                            rebuilding: true,
                                        });
                                        this.props.onRebuildClick(this.props.submission_id);
                                        
                                    }}>
                                        {this.state.rebuilding ? "Rebuilding" : "Rebuild"}
                                </button>
                            </p>
                            {approveButton}
                        </div>
                    </Row>
                </div>
            </Row>
        );
    }

    private handleClick(rebuild: () => void) {
        // TODO: implement rebuild functionality
        rebuild();
        console.log("Rebuilding...");
    }
}
