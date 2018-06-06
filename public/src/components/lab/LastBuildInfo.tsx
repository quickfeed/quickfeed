import * as React from "react";
import { Row } from "../../components";

interface ILastBuildInfo {
    pass_tests: number;
    fail_tests: number;
    exec_time: number;
    build_time: Date;
    build_id: number;
}
class LastBuildInfo extends React.Component<ILastBuildInfo, any> {

    public render() {
        return (
            <Row>
                <div className="col-lg-8">
                    <h2>Latest build</h2>
                    <p id="passes">Number of passed tests:  {this.props.pass_tests}</p>
                    <p id="fails">Number of failed tests:  {this.props.fail_tests}</p>
                    <p id="buildtime">Execution time:  {this.props.exec_time / 1000} s</p>
                    <p id="timedate">Build date:  {this.props.build_time.toString()}</p>
                    <p id="buildid">Build ID: {this.props.build_id}</p>
                </div>
                <div className="col-lg-4 hidden-print">
                    <h2>Actions</h2>
                    <Row>
                        <div className="col-lg-12">
                            <p>
                                <button type="button" id="rebuild" className="btn btn-primary"
                                    onClick={() => this.handleClick()}>Rebuild
                                </button>
                            </p>
                        </div>
                    </Row>
                </div>
            </Row>
        );
    }

    private handleClick() {
        // TODO: implement rebuild functionality
        console.log("Rebuilding...");
    }
}
export { LastBuildInfo };
