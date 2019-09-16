import * as React from "react";
import { ProgressBar, Row } from "../../components";

interface ILabResult {
    progress: number;
    lab: string;
    course_name: string;
    authorName?: string;
    isApproved: boolean;
    delivered: string;
}

interface ILabResultState {
    approved: boolean;
}

export class LabResult extends React.Component<ILabResult, ILabResultState> {

    constructor(props: ILabResult) {
        super(props);
        this.state = {
            approved: this.props.isApproved,
        };
    }

    public render() {
        let labHeading: JSX.Element;
        if (this.props.authorName) {
            console.log("Author name is " + this.props.authorName);
            labHeading = <h3>{this.props.authorName + ": "} {this.props.lab}</h3>;
        } else {
            labHeading = <div>
                <p className="lead">Your progress on <strong><span
                    id="lab-headline">{this.props.lab}</span></strong>
                </p>
            </div>;
        }
        console.log("Progress is " + this.props.progress
         + " and status is " + this.setApprovedString() + " delivered on " + this.props.delivered);
        return (
            <Row>
                <div className="col-lg-12">
                    {labHeading}
                    <ProgressBar progress={this.props.progress}></ProgressBar>
                </div>
                <div className="col-lg-6">
                    <p><strong id="status">Status: {this.setApprovedString()}</strong></p>
                </div>
                <div className="col-lg-6">
                    <p><strong id="pushtime">Delivered: {this.props.delivered} </strong></p>
                </div>
            </Row>
        );
    }

    private setApprovedString(): string {
        return this.props.isApproved ? "Approved" : "Not approved";
    }
}
