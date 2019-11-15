import * as React from "react";
import { Row } from "../../components";

interface ILastBuildInfo {
    submission_id: number;
    pass_tests: number;
    fail_tests: number;
    exec_time: number;
    build_time: string;
    isApproved: boolean;
    showApprove: boolean;
    onApproveClick: () => void;
    onRebuildClick: (submissionID: number) => Promise<boolean>;
}

interface ILastBuildInfoState {
    rebuilding: boolean;
}

export class LastBuildInfo extends React.Component<ILastBuildInfo, ILastBuildInfoState> {

    constructor(props: ILastBuildInfo) {
        super(props);
        this.state = {
            rebuilding: false,
         };
    }
    public render() {
        return (
            <Row>
                <div className="col-lg-8">
                    <h2>Latest build</h2>
                    <p id="passes">Passed tests:  {this.props.pass_tests}</p>
                    <p id="fails">Failed tests:  {this.props.fail_tests}</p>
                    <p id="buildtime">Execution time:  {this.props.exec_time / 1000} s</p>
                    <p id="timedate">Build date:  {this.props.build_time ? this.props.build_time.toString() : "-"}</p>
                </div>
            </Row>
        );
    }

}
