import * as React from "react";
import { DynamicTable, Row } from "../../components";
import { Score } from "../../../proto/kit/score/score_pb";

import { ICellElement } from "../data/DynamicTable";

interface ILastBuildProps {
    test_cases: Score[];
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
                        selector={(sc: Score) => [sc.getTestname(),
                         (sc.getScore() ? sc.getScore().toString() : "0")
                          + "/" + (sc.getMaxscore() ? sc.getMaxscore().toString() : "0") + " pts",
                          sc.getWeight() ? sc.getWeight().toString() : "0"]}
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
