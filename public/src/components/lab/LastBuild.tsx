import * as React from "react";
import { DynamicTable, Row } from "../../components";
import { ITestCases } from "../../models";

interface ILastBuild {
    test_cases: ITestCases[];
    score: number;
    weight: number;
}
class LastBuild extends React.Component<ILastBuild, any> {

    public render() {
        return (
            <Row>
                <div className="col-lg-12">
                    <DynamicTable
                        header={["Test name", "Score", "Weight"]}
                        data={this.props.test_cases}
                        selector={(item: ITestCases) => [item.name, item.score.toString() + "/"
                            + item.points.toString() + " pts", item.weight.toString() + " pts"]}
                        footer={["Total score", this.props.score.toString() + "%", this.props.weight.toString() + "%"]}
                    />
                </div>
            </Row>
        );
    }
}
export { LastBuild };
