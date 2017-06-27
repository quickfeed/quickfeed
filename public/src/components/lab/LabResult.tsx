import * as React from "react";
import { ProgressBar, Row } from "../../components";

interface ILabResult {
    progress: number;
    lab: string;
    course_name: string;
}
class LabResult extends React.Component<ILabResult, any> {

    public render() {
        return (
            <Row>
                <div className="col-lg-12">
                    <h1>{this.props.course_name}</h1>
                    <p className="lead">Your progress on <strong><span
                        id="lab-headline">{this.props.lab}</span></strong>
                    </p>
                    <ProgressBar progress={this.props.progress}></ProgressBar>
                </div>
                <div className="col-lg-6">
                    <p><strong id="status">Status: Nothing built yet.</strong></p>
                </div>
                <div className="col-lg-6">
                    <p><strong id="pushtime">Code delievered: - </strong></p>
                </div>
            </Row>
        );
    }
}

export { LabResult };
