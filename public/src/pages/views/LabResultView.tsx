import * as React from "react";
import { LabResult, LastBuild, LastBuildInfo, Row } from "../../components";
import { ICourse, IStudentSubmission, ISubmission } from "../../models";

interface ILabInfoProps {
    course: ICourse;
    labInfo: IStudentSubmission;
}

export class LabResultView extends React.Component<ILabInfoProps, {}> {

    public render() {
        if (this.props.labInfo.latest) {
            return (
                <div className="col-md-9 col-sm-9 col-xs-12">
                    <div className="result-content" id="resultview">
                        <section id="result">
                            <LabResult
                                course_name={this.props.course.name}
                                lab={this.props.labInfo.assignment.name}
                                progress={this.props.labInfo.latest.score}
                            />
                            <LastBuild
                                test_cases={this.props.labInfo.latest.testCases}
                                score={this.props.labInfo.latest.score}
                                weight={100}
                            />
                            <LastBuildInfo
                                pass_tests={this.props.labInfo.latest.passedTests}
                                fail_tests={this.props.labInfo.latest.failedTests}
                                exec_time={this.props.labInfo.latest.executetionTime}
                                build_time={this.props.labInfo.latest.buildDate}
                                build_id={this.props.labInfo.latest.buildId}
                            />
                            <Row>
                                <div className="col-lg-12">
                                    <div className="well">
                                        <code id="logs">{this.props.labInfo.latest.buildLog}</code>
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
