import * as React from "react";
import { ProgressBar, Row } from "../../components";
import { IUser } from "../../models";

interface ILabResult {
    progress: number;
    lab: string;
    course_name: string;
    student?: IUser;
}
class LabResult extends React.Component<ILabResult, any> {

    public render() {
        let labHeading: JSX.Element;
        if (this.props.student) {
            labHeading = <h3>{this.props.student.firstname + " " + this.props.student.lastname}: {this.props.lab}</h3>;
        } else {
            labHeading = <div>
                <h1>{this.props.course_name}</h1>
                <p className="lead">Your progress on <strong><span
                    id="lab-headline">{this.props.lab}</span></strong>
                </p>
            </div>;
        }
        return (
            <Row>
                <div className="col-lg-12">
                    {labHeading}
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
