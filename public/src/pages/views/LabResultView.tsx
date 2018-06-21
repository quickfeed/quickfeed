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

export class LabResultView extends React.Component<ILabInfoProps, {}> {

    public render() {
        if (this.props.labInfo.latest) {
            const latest = this.props.labInfo.latest;
            console.log(latest.buildLog);
            const buildLog = latest.buildLog.split("\n").map(x => <span>{x}<br /></span>);

            return (
                <div className="col-md-9 col-sm-9 col-xs-12">
                    <div className="result-content" id="resultview">
                        <section id="result">
                            <LabResult
                                course_name={this.props.course.name}
                                lab={this.props.labInfo.assignment.name}
                                progress={latest.score}
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
}
