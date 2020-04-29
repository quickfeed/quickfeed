import * as React from "react";
import { GradingCriterion } from "../../../proto/ag_pb";

interface EditCriterionProps {
    criterion: GradingCriterion;
    onUpdate: (newDescription: string) => void;
    onDelete: () => void;

    // assignment?: boolean // editable description if assignment view, editable passed/failed if not
}

interface EditCriterionState {
    editing: boolean;
    description: string;
}

export class EditCriterion extends React.Component<EditCriterionProps, EditCriterionState> {

    constructor(props: EditCriterionProps) {
        super(props);
        this.state = {
            editing: false,
            description: this.props.criterion.getDescription(),
        }
    }

    public render() {
        return <div className="c-element">
            {this.state.editing ? this.renderEditView() : this.renderTextView()}
        </div>;
    }

    private renderDeleteButton(): JSX.Element {
        return <button
            className="btn btn-danger btn-xs"
            onClick={() => this.props.onDelete()}
        >X</button>
    }

    private toggleEditState() {
        this.setState({
            editing: !this.state.editing,
        })
    }

    private updateDescription() {
        this.props.onUpdate(this.state.description);
        this.setState({
            editing: false,
            description: this.props.criterion.getDescription(),
        });
    }

    private renderTextView(): JSX.Element {
        return <div
            onDoubleClick={() => this.toggleEditState()}
    >{this.props.criterion.getDescription()}{this.renderDeleteButton()}</div>
    }

    private renderEditView(): JSX.Element {
        return <div className="input-btns">
            <input
                type="text"
                defaultValue={this.state.description}
                onChange={(e) => this.setDescription(e.target.value)}
        />
        <div className="btn-group">
        <button
            className="btn btn-primary btn-xs"
            onClick={() => this.updateDescription()}>OK</button>
        <button
            className="btn btn-danger btn-xs"
            onClick={() => this.setState({editing: false, description: this.props.criterion.getDescription()})}>X</button></div>
        </div>
    }

    private setDescription(inputText: string) {
        this.setState({
            description: inputText,
        })
    }

    // reset state: editing false, desc to initial
}