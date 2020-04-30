import * as React from "react";
import { GradingCriterion, Status } from '../../../proto/ag_pb';

interface GradeCriterionProps {
    criterion: GradingCriterion;
    addComment: (comment: string) => void;
    addGrade: (grade: GradingCriterion.Grade) => void;
}

interface GradeCriterionState {
    grade: GradingCriterion.Grade;
    comment: string;
    commenting: boolean;
}

export class GradeCriterion extends React.Component<GradeCriterionProps, GradeCriterionState> {

    constructor(props: GradeCriterionProps) {
        super(props);
        this.state = {
            grade: this.props.criterion.getGrade(),
            comment: this.props.criterion.getComment(),
            commenting: false,
        }
    }

    public render() {
        return <div className="c-element">
            {this.props.criterion.getDescription()}{this.renderSwitch()}
            {this.renderComment()}
        </div>
    }

    private renderSwitch() {
        return <div className="switch-toggle">
        <input id="on" name="state-d" type="radio" checked={false} />
        <label>ON</label>

        <input id="na" name="state-d" type="radio" disabled checked={true} />
        <label className="disabled">&nbsp;</label>

        <input id="off" name="state-d" type="radio" />
        <label>OFF</label>
      </div>

      // TODO: update grade locally, only update database when prompted by TA
    }

    private renderComment(): JSX.Element {
        const commentDiv = <div className="comment-div"
            onDoubleClick={() => this.toggleEdit()}
            >{this.state.comment}</div>;
        const editDiv = <div className="input-group">
            <input
                type="text"
                defaultValue={this.state.comment}
                onChange={(e) => this.setComment(e.target.value)}
                onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                        this.updateComment();
                    }
                }}
            />
            <button
                className="btn btn-primary btn-xs"
                onClick={() => this.updateComment()}>OK</button>
            <button
                className="btn btn-danger btn-xs"
                onClick={() => this.toggleEdit()}>X</button></div>
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
            comment: this.props.criterion.getComment(),
        })
    }

    private toggleEdit() {
        this.setState({
            commenting: !this.state.commenting,
        })
    }
}