import * as React from "react";
import { Course } from "../../../proto/ag_pb";
import { LabResult, LastBuild, LastBuildInfo, Row } from "../../components";
import { IStudentSubmission } from "../../models";

interface ILabInfoProps {
    course: Course;
    labInfo: IStudentSubmission;
    showApprove: boolean;
    authorName?: string;
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
                                course_name={this.props.course.getName()}
                                lab={this.props.labInfo.assignment.getName()}
                                progress={latest.score}
                                isApproved={latest.approved}
                                authorName={this.props.authorName}
                                delivered={this.getDeliveredTime(latest.buildDate)}
                            />
                            <LastBuild
                                test_cases={latest.testCases}
                                score={latest.score}
                                weight={100}
                            />
                            <LastBuildInfo
                                submission_id={latest.id}
                                pass_tests={latest.passedTests}
                                fail_tests={latest.failedTests}
                                exec_time={latest.executetionTime}
                                build_time={this.getDeliveredTime(latest.buildDate)}
                                build_id={latest.buildId}
                                isApproved={latest.approved}
                                onApproveClick={this.props.onApproveClick}
                                onRebuildClick={this.props.onRebuildClick}
                                showApprove={this.props.showApprove}
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

    private getDeliveredTime(date: Date): string {
        return date ? date.toDateString() : "-";
    }
}
