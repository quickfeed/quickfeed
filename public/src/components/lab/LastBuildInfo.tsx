import * as React from "react";
import { Assignment, Submission } from "../../../proto/ag/ag_pb";
import { Row } from "../../components";
import { formatDate } from "../../helper";
import { ISubmission } from "../../models";
import { submissionStatusToString, getDaysAfterDeadline } from "../../componentHelper";

interface ILastBuildInfoProps {
    submission: ISubmission;
    assignment: Assignment;
    teacherView: boolean;
    slipdays: number;
}

interface ILastBuildInfoState {
    rebuilding: boolean;
}

export class LastBuildInfo extends React.Component<ILastBuildInfoProps, ILastBuildInfoState> {
    constructor(props: ILastBuildInfoProps) {
        super(props);
        this.state = {
            rebuilding: false,
        };
    }

    public render() {
        const alltests = this.props.submission.testCases ? this.props.submission.testCases.length : 0;
        const passedAllTests = this.props.submission.passedTests === alltests ? "passing" : "";
        const slipDaysRow = <tr><td key="5">Slip days</td><td key="desc5">{this.props.slipdays}</td></tr>;
        const approvedLine = <tr><td key="6">Approved</td><td key="desc6">{formatDate(this.props.submission.approvedDate)}</td></tr>

        return (
            <div>
                <Row>
                    <div className="col-lg-12">
                        <table className="table">
                            <thead key="thead"><tr><th key="headrow" colSpan={2}>Lab Information </th></tr></thead>
                            <tbody key="tbody">
                                <tr><td key="status">Status</td><td key="desc_0">{this.setStatusString()}</td></tr>
                                <tr><td key="1">Delivered</td><td key="desc1">{this.getDeliveredTime()}</td></tr>
                                {this.showApprovedLine() ? approvedLine : null}
                                <tr><td key="2">Deadline</td><td key="desc2">{formatDate(this.props.assignment.getDeadline())}</td></tr>
                                <tr><td key="3">Tests passed</td><td key="desc3"><div className={passedAllTests}>{this.props.submission.passedTests} / {alltests}</div></td></tr>
                                <tr><td key="4">Execution time</td><td key="desc4">{this.formatTime(this.props.submission.buildInfo.getExectime())} seconds </td></tr>
                                {this.props.assignment.getIsgrouplab() ? null : slipDaysRow}
                                </tbody>
                        </table>
                    </div>
                </Row>
            </div>
        );
    }

    private showApprovedLine(): boolean {
        return this.props.submission.status === Submission.Status.APPROVED && this.props.submission.approvedDate !== "";
    }

    private getDeliveredTime(): JSX.Element {
        const deadline = new Date(this.props.assignment.getDeadline());
        const delivered = this.props.submission.buildDate;
        const showAfterDeadline = this.props.teacherView && this.props.submission.status !== Submission.Status.APPROVED
        let classString = "";
        if (delivered >= deadline && showAfterDeadline) {
            classString = "past-deadline";
        }
        return <div className={classString}>{this.formatDelivered(deadline, delivered, showAfterDeadline)}</div>;
    }

    private formatDelivered(deadline: Date, delivered: Date, showAfterDeadline: boolean): string {
        const daysAfter = getDaysAfterDeadline(deadline, delivered);
        const daysAfterString = daysAfter > 0 ? " (" + daysAfter + " days after deadline)" : "";
        return formatDate(delivered) + (showAfterDeadline ? daysAfterString : "");
    }

    private formatTime(executionTime: number): number {
        return executionTime / 1000.0;
    }

    private setStatusString(): JSX.Element {
        const approved = this.props.submission.status === Submission.Status.APPROVED
        const className = approved ? "greentext" : "";
        if (this.props.assignment.getReviewers() > 0) {
            return approved ? <div className="greentext">Approved</div> : <div>{submissionStatusToString(this.props.submission.status)}</div>
        }
        return <div className={className}>{submissionStatusToString(this.props.submission.status)}</div>;
    }

}
