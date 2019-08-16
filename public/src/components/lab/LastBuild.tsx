import * as React from "react";
import { DynamicTable, Row } from "../../components";
import { ITestCases } from "../../models";

interface ILastBuild {
    test_cases: ITestCases[];
    score: number;
    weight: number;
}

export class LastBuild extends React.Component<ILastBuild> {

    public render() {
        return (
            <Row>
                <div className="col-lg-12">
                    <DynamicTable
                        header={["Test name", "Score", "Weight"]}
                        data={this.props.test_cases}
                        selector={(item: ITestCases) => [item.name, item.score ? item.score.toString() : "0" + "/"
                            + item.points ? item.points.toString() : "0" + " pts", item.weight ? item.weight.toString()
                            : "0" + " pts"]}
                        footer={["Total score", this.props.score ? this.props.toString()
                        : "0" + "%", this.props.weight ? this.props.weight.toString() : "0" + "%"]}
                    />
                </div>
            </Row>
        );
    }
}
