import * as React from "react";
import { LabResult, LastBuild, LastBuildInfo, Row } from "../../components";
import { ICourse, IStudentSubmission, ISubmission } from "../../models";

interface ILabInfoProps {
    course: ICourse;
    labInfo: IStudentSubmission;
    showApprove: boolean;
    onApproveClick: () => void;
    onRebuildClick: () => void;
}

function isDate(date: any): date is Date {
    return (date as any).getDate !== undefined;
}

export class LabResultView extends React.Component<ILabInfoProps, {}> {

    public render() {
        if (this.props.labInfo.latest) {
            const latest = this.props.labInfo.latest;
            const buildLog = latest.buildLog.split("\n").map((x) => <span>{x}<br /></span>);

            return (
                <div className="col-md-9 col-sm-9 col-xs-12">
                    <div className="result-content" id="resultview">
                        <section id="result">
                            <LabResult
                                course_name={this.props.course.name}
                                lab={this.props.labInfo.assignment.name}
                                progress={latest.score}
                                status={this.getSubmissionInfo()}
                                deliverd={this.getCodeDeliverdString(this.props.labInfo.latest.buildDate)}
                            />
                            <LastBuild
                                test_cases={latest.testCases}
                                score={latest.score}
                                weight={100}
                            />
                            <LastBuildInfo
                                pass_tests={latest.passedTests}
                                fail_tests={latest.failedTests}
                                exec_time={latest.executetionTime}
                                build_time={latest.buildDate}
                                build_id={latest.buildId}
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
        return <h1>No subissions have been submitted yet</h1>;
    }

    private getSubmissionInfo(): string {
        if (this.props.labInfo.latest) {
            return this.props.labInfo.latest.approved ? "Approved" : "Not approved";
        }
        return "Nothing built yet!";
    }

    private getCodeDeliverdString(date?: Date | string): string {
        if (date && isDate(date)) {
            return date.toDateString();
        } else if (typeof (date) === "string") {
            return date;
        }
        return "-";
    }
}
