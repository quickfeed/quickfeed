import * as React from "react";
import { DynamicTable, Row } from "../../components";
import { ITestCases } from "../../models";
import { ICellElement } from "../data/DynamicTable";

interface ILastBuildProps {
    test_cases: ITestCases[];
    score: number;
    weight: number;
    scoreLimit: number;
}

export class LastBuild extends React.Component<ILastBuildProps> {

    public render() {
        return (
            <Row>
                <div className="col-lg-12">
                    <DynamicTable
                        header={["Test name", "Score", "Weight"]}
                        data={this.props.test_cases ? this.props.test_cases : []}
                        selector={(item: ITestCases) => [item.TestName ? item.TestName : "-",
                         (item.Score ? item.Score.toString() : "0")
                          + "/" + (item.MaxScore ? item.MaxScore.toString() : "0") + " pts",
                          item.Weight ? item.Weight.toString() : "0"]}
                        footer={this.makeDynamicFooter()}
                    />
                </div>
            </Row>
        );
    }

    private makeDynamicFooter(): ICellElement[] {
        return [
            {value: "Total score"},
            this.makeScoreCell(this.props.score, this.props.scoreLimit),
            {value: this.props.weight ? this.props.weight.toString() + " %" : "-"},
        ];
    }

    private makeScoreCell(score: number, scoreLimit: number): ICellElement {
        const cellClass = ((scoreLimit > 0) && (score >= scoreLimit)) ? "passing" : "test";
        return {value: score.toString() + " %", className: cellClass};
    }
}
