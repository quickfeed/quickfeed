import * as React from "react";
import { Course } from "../../../proto/ag_pb";
import { LabResult, LastBuild, LastBuildInfo, Row } from "../../components";
import { IStudentSubmission } from "../../models";

interface ILabInfoProps {
    course: Course;
    labInfo: IStudentSubmission;
    showApprove: boolean;
    onApproveClick: () => void;
    onRebuildClick: (submissionID: number) => Promise<boolean>;
}

export class LabResultView extends React.Component<ILabInfoProps> {

    public render() {
        if (this.props.labInfo.latest) {
            const latest = this.props.labInfo.latest;
            const buildLog = latest.buildLog.split("\n").map((x) => <span>{x}<br /></span>);
            return (
                <div className="col-md-9 col-sm-9 col-xs-12">
                    <div className="result-content" id="resultview">
                        <section id="result">
                            <LabResult
                                submission_id={latest.id}
                                showApprove={this.props.showApprove}
                                lab={this.props.labInfo.assignment.getName()}
                                progress={latest.score}
                                isApproved={latest.approved}
                                authorName={this.props.labInfo.authorName}
                                onApproveClick={this.props.onApproveClick}
                                onRebuildClick={this.props.onRebuildClick}
                            />
                            <LastBuildInfo
                                submission={latest}
                                assignment={this.props.labInfo.assignment}
                            />
                            <LastBuild
                                test_cases={latest.testCases}
                                score={latest.score}
                                weight={100}
                            />
                            <Row>
                                <div className="col-lg-12">
                                    <div className="well">
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

    private getSubmissionInfo(): string {
        if (this.props.labInfo.latest) {
            return this.props.labInfo.latest.approved ? "Approved" : "Not approved";
        }
        return "Nothing built yet!";
    }

    
}
