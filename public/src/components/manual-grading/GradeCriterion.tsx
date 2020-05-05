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
        return <div className="switch-toggle btn-group">
            <button className={this.setButtonClass(GradingCriterion.Grade.PASSED, "btn-success")}
                onClick={() => {
                    this.props.addGrade(GradingCriterion.Grade.PASSED);
                    this.setState({
                        grade: GradingCriterion.Grade.PASSED,
                    });
                }}
            ><span className="glyphicon glyphicon-ok-circle"></span></button>
            <button className={this.setButtonClass(GradingCriterion.Grade.NONE, "btn-basic")}
                onClick={() => {
                    this.props.addGrade(GradingCriterion.Grade.NONE);
                    this.setState({
                        grade: GradingCriterion.Grade.NONE,
                    });
                }}
            ><span className="glyphicon glyphicon-ban-circle"></span></button>
            <button className={this.setButtonClass(GradingCriterion.Grade.FAILED, "btn-danger")}
                onClick={() => {
                    this.props.addGrade(GradingCriterion.Grade.FAILED);
                    this.setState({
                        grade: GradingCriterion.Grade.FAILED,
                    });
                }}
            ><span className="glyphicon glyphicon-remove-circle"></span></button>
        </div>
      // TODO: update grade locally, only update database when prompted by TA
    }

    private setButtonClass(grade: GradingCriterion.Grade, classString: string): string {
        return "btn btn-xs " + (this.state.grade === grade ? classString : "btn-default");
    }

    private renderComment(): JSX.Element {
        const commentDiv = <div className="comment-div"
            onDoubleClick={() => this.toggleEdit()}
            >{this.state.comment !== "" ? this.state.comment : "Add new comment"}</div>;
        const editDiv = <div className="input-group">
            <input
                autoFocus={true}
                type="text"
                defaultValue={this.state.comment}
                onChange={(e) => this.setComment(e.target.value)}
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
            comment: this.props.criterion.getComment(),
        })
    }

    private toggleEdit() {
        this.setState({
            commenting: !this.state.commenting,
        })
    }
}