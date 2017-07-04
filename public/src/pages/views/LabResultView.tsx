import * as React from "react";
import {LabResult, LastBuild, LastBuildInfo, Row} from "../../components";
import {ILabInfo} from "../../models";

interface ILabInfoProps {
    labInfo: ILabInfo;
}
class LabResultView extends React.Component<ILabInfoProps, {}> {

    public render() {
        return (
            <div className="col-md-9 col-sm-9 col-xs-12">
                <div className="result-content" id="resultview">
                    <section id="result">
                        <LabResult
                            course_name={this.props.labInfo.course}
                            lab={this.props.labInfo.lab}
                            progress={this.props.labInfo.score}
                            student={this.props.labInfo.student}
                        />
                        <LastBuild
                            test_cases={this.props.labInfo.test_cases}
                            score={this.props.labInfo.score}
                            weight={this.props.labInfo.weight}
                        />
                        <LastBuildInfo
                            pass_tests={this.props.labInfo.pass_tests}
                            fail_tests={this.props.labInfo.fail_tests}
                            exec_time={this.props.labInfo.exec_time}
                            build_time={this.props.labInfo.build_time}
                            build_id={this.props.labInfo.build_id}
                        />
                        <Row>
                            <div className="col-lg-12">
                                <div className="well">
                                    <code id="logs"># There is no build for this lab yet.</code>
                                </div>
                            </div>
                        </Row>

                    </section>
                </div>
            </div>
        );
    }
}

export {LabResultView};
