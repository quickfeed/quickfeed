import * as React from "react";
import { GradingCriterion } from "../../../proto/ag_pb";

interface EditCriterionProps {
    criterion: GradingCriterion;
    onUpdate: (newDescription: string) => void;
    onDelete: () => void;
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
            className="btn btn-danger btn-xs bm-btn"
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
        this.undo();
    }

    private renderTextView(): JSX.Element {
        return <div
            onDoubleClick={() => this.toggleEditState()}
    >{this.props.criterion.getDescription()}{this.renderDeleteButton()}</div>
    }

    private renderEditView(): JSX.Element {
        return <div className="input-group">
            <input
                autoFocus={true}
                type="text"
                defaultValue={this.state.description}
                onChange={(e) => this.setDescription(e.target.value)}
                onBlur={() => this.undo()}
                onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                        this.updateDescription();
                    } else if (e.key === 'Escape') {
                        this.undo();
                    }
                }}
        /></div>
    }

    private setDescription(inputText: string) {
        this.setState({
            description: inputText,
        })
    }

    private undo() {
        this.setState({
            editing: false,
            description: this.props.criterion.getDescription(),
        })
    }
}