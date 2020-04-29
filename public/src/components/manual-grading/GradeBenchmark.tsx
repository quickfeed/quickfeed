import * as React from "react";
import { GradingBenchmark, GradingCriterion, Grade } from '../../../proto/ag_pb';
import { GradeCriterion } from "./GradeCriterion";

interface GradeBenchmarkProps {
    benchmark: GradingBenchmark,
    addComment: (comment: string) => void;
}

interface GradeBenchmarkState {
    criteria: GradingCriterion[];
    commenting: boolean;
    comment: string;
}

export class GradeBenchmark extends React.Component<GradeBenchmarkProps, GradeBenchmarkState> {
    constructor(props: GradeBenchmarkProps) {
        super(props);
        this.state = {
            criteria: this.props.benchmark.getCriteriaList(),
            comment: this.props.benchmark.getComment(),
            commenting: false,
        }
    }

    public render() {
        return <div>
            <h3 className="b-header">{this.props.benchmark.getHeading()}</h3>
            {this.renderComment()}
            {this.renderList()}
        </div>
    }


    private renderList(): JSX.Element[] {
        return this.state.criteria.map((c, i) => <GradeCriterion
            criterion={c}
            addComment={(comment: string) => {
                c.setComment(comment);
            }}
            addGrade={(grade: GradingCriterion.Grade) => {
                c.setGrade(grade);
            }}
        />)
    }

    private renderComment(): JSX.Element {
        // TODO
        return <div></div>
    }

    private setComment(input: string) {
        this.setState({
            comment: input,
        });
    }

    private updateComment() {
        this.props.addComment(this.state.comment);
        this.setState({
            commenting: false,
        });
    }

    private toggleEdit() {
        this.setState({
            commenting: !this.state.commenting,
        });
    }
}