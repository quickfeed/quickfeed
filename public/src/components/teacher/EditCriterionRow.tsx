import * as React from "react";
import { GradingCriterion } from '../../../proto/ag_pb';

interface EditCriterionRowProps {
    criterion: GradingCriterion;
    onUpdate: (criterion: GradingCriterion) => boolean;
    onDelete: () => void;

    // assignment?: Assignment // editable description if assignment view, editable passed/failed if not
}

interface EditCriterionRowState {
    editing: boolean;
    description: string;
}

export class EditCriterionRow extends React.Component<EditCriterionRowProps, EditCriterionRowState> {


    constructor(props: EditCriterionRowProps) {
        super(props);
        this.state = {
            editing: false,
            description: this.props.criterion.getDescription(),
        }
    }

    public render() {
        return this.state.editing ?
        this.renderEditView() : this.renderTextView();
    }

    private toggleEditState() {
        this.setState({
            editing: !this.state.editing,
        })
    }

    private updateDescription() {
        this.setState({
            editing: false,
        }, () => {
            const c = this.props.criterion;
            c.setDescription(this.state.description);
            if (!this.props.onUpdate(c)) {
                this.setState({
                    description: this.props.criterion.getDescription(),
                })
            }

        })
    }

    private renderTextView(): JSX.Element {
        return <div
            onDoubleClick={() => this.toggleEditState()}
        >{this.props.criterion.getDescription()}</div>
    }

    private renderEditView(): JSX.Element {
        return <div>
            <input
                type="text"
                defaultValue={this.state.description}
                onChange={(e) => this.setDescription(e.target.value)}
        />
        <div className="btn-group action-btn">
        <button
            className="btn btn-primary"
            onClick={() => this.updateDescription()}>OK</button>
        <button
            className="btn btn-danger"
            onClick={() => this.props.onDelete()}>X</button></div>
        </div>
    }

    private setDescription(inputText: string) {
        this.setState({
            description: inputText,
        })
    }
}