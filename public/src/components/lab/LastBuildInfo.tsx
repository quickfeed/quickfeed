import * as React from "react";
import { Assignment } from "../../../proto/ag_pb";
import { Row } from "../../components"
import { ISubmission } from "../../models";
import { string } from "prop-types";

interface ILastBuildInfo {
    submission: ISubmission;
    assignment: Assignment;
    }

interface ILastBuildInfoState {
    rebuilding: boolean;
}

export class LastBuildInfo extends React.Component<ILastBuildInfo, ILastBuildInfoState> {

    private options = {
        weekday: "short",
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "numeric",
        hour12: false,
    }
    constructor(props: ILastBuildInfo) {
        super(props);
        this.state = {
            rebuilding: false,
         };
    }


    public render() {
        const deadline = new Date(this.props.assignment.getDeadline());
        return (
            <div>
                <Row>
                <div className="col-lg-12">
                    <table className="table">
                        <thead><tr><th colSpan={2}>Lab Information </th></tr></thead>
                        <tbody>
        <tr><td>Delivered</td><td>{this.getDeliveredTime()}</td></tr>
    <tr><td>Deadline</td><td>{deadline.toLocaleString("no-NO", this.options)}</td></tr>
                            <tr><td>Tests passed</td><td>{this.props.submission.passedTests} / {this.props.submission.testCases.length}</td></tr>
        <tr><td>Execution time</td><td>{this.props.submission.executetionTime / 1000} s</td></tr>
        <tr><td>Slip days</td><td>5</td></tr>
                        </tbody>
                    </table>
                </div>
            </Row>
            </div>
        );
    }

    private getDeliveredTime(): JSX.Element {
        const deadline = new Date(this.props.assignment.getDeadline());
        const delivered = this.props.submission.buildDate;
        let classString = "";
        if (delivered >= deadline) {
            classString = "past-deadline";
        }

        return <div className={classString}>{delivered.toLocaleString("no-NO", this.options)}</div>;
    }

}
