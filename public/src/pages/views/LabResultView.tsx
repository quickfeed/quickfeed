import * as React from "react";
import { LabResult, LastBuild, LastBuildInfo, Row } from "../../components";
import { IStudentSubmission } from "../../models";

interface ILabInfoProps {
    assignment: IStudentSubmission;
    showApprove: boolean;
    onApproveClick: (approve: boolean) => void;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<boolean>;
}

export class LabResultView extends React.Component<ILabInfoProps> {

    public render() {
        if (this.props.assignment.latest) {
            const latest = this.props.assignment.latest;
            const buildLog = latest.buildLog.split("\n").map((x, i) => <span key={i} >{x}<br /></span>);
            return (
                <div key="labhead" className="col-md-9 col-sm-9 col-xs-12">
                    <div key="labview" className="result-content" id="resultview">
                        <section id="result">
                            <LabResult
                                assignment_id={this.props.assignment.assignment.getId()}
                                submission_id={latest.id}
                                showApprove={this.props.showApprove}
                                lab={this.props.assignment.assignment.getName()}
                                progress={latest.score}
                                isApproved={latest.approved}
                                authorName={this.props.assignment.authorName}
                                onApproveClick={this.props.onApproveClick}
                                onRebuildClick={this.props.onRebuildClick}
                            />
                            <LastBuildInfo
                                submission={latest}
                                assignment={this.props.assignment.assignment}
                            />
                            <LastBuild
                                test_cases={latest.testCases}
                                score={latest.score}
                                scoreLimit={this.props.assignment.assignment.getScorelimit()}
                                weight={100}
                            />
                            <Row>
                                <div key="loghead" className="col-lg-12">
                                    <div key="logview" className="well">
                                        <code id="logs">{buildLog}</code>
                                    </div>
                                </div>
                            </Row>
                        </section>
                    </div>
                </div>
            );
        }
        return <h1>No submissions yet</h1>;
    }
}
