import * as React from "react";
import { GradingBenchmark, GradingCriterion } from '../../../proto/ag_pb';
import { GradeCriterion } from "./GradeCriterion";

interface GradeBenchmarkProps {
    benchmark: GradingBenchmark,
    addComment: (comment: string) => void;
    onUpdate: (criteria: GradingCriterion[]) => void;
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
            <h3 className="b-header">{this.props.benchmark.getHeading()}<span className="glyphicon glyphicaon-comment"></span></h3>
            {this.renderComment()}
            {this.renderList()}
        </div>
    }

    private renderList(): JSX.Element[] {
        return this.state.criteria.map((c, i) => <GradeCriterion
            key={"c" + i}
            criterion={c}
            addComment={(comment: string) => {
                c.setComment(comment);
                this.props.onUpdate(this.state.criteria);
            }}
            addGrade={(grade: GradingCriterion.Grade) => {
                c.setGrade(grade);
                this.props.onUpdate(this.state.criteria);
            }}
        />)
    }

    private renderComment(): JSX.Element {
        const commentDiv = <div className="comment-div col-md-12"
            onClick={() => this.toggleEdit()}
            >{this.state.comment !== "" ? this.state.comment : "Add a benchmark comment"}</div>;
        const editDiv = <div className="input-group col-md-12">
            <input
                className="form-control m-input"
                autoFocus={true}
                type="text"
                defaultValue={this.state.comment}
                onChange={(e) => this.setComment(e.target.value)}
                onBlur={() => this.toggleEdit()}
                onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                        this.updateComment();
                    } else if (e.key === 'Escape') {
                        this.toggleEdit();
                    }
                }}
            /></div>
        return <div className="comment-div">
            {this.state.commenting ? editDiv : commentDiv}
        </div>
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
            comment: this.props.benchmark.getComment(),
        });
    }

    private toggleEdit() {
        this.setState({
            commenting: !this.state.commenting,
        });
    }
}