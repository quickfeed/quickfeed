import * as React from "react";
import { GradingBenchmark, GradingCriterion } from "../../../proto/ag_pb";
import { EditCriterion } from "./EditCriterion";

interface EditBenchmarkProps {
    customScore: boolean;
    benchmark: GradingBenchmark,
    onAdd: (c: GradingCriterion) => Promise<GradingCriterion | null>;
    onUpdate: (newHeading: string) => void;
    onDelete: () => void;

    updateCriterion: (c: GradingCriterion) => Promise<boolean>;
    deleteCriterion: (c: GradingCriterion) => Promise<boolean>;
}

interface EditBenchmarkState {
    editing: boolean;
    adding: boolean;
    heading: string;
    criteria: GradingCriterion[];
    newCriterion: string;
}

export class EditBenchmark extends React.Component<EditBenchmarkProps, EditBenchmarkState> {

    constructor(props: EditBenchmarkProps) {
        super(props);
        this.state = {
            editing: false,
            adding: false,
            heading: this.props.benchmark.getHeading(),
            criteria: this.props.benchmark.getCriteriaList(),
            newCriterion: "",
        }
    }

    public render() {
        return <div className="b-element">
            <h3 className="b-header" onClick={() => this.toggleEdit()}>
                {this.state.editing ? this.renderEditView() : this.renderTextView()}
            </h3>

        {this.renderCriteriaList()}

        {this.renderAddRow() }
        </div>
    }



    private removeButton(): JSX.Element {
        return <button className="btn btn-danger btn-xs bm-btn" onClick={
            () => {
                this.setState({
                    editing: false,
                })
                this.props.onDelete();
            }
        }>X</button>
    }

    private renderAddRow(): JSX.Element {
        const addDiv = <div className="row c-add-row" onClick={() => this.toggleAdd()}>
    <span className="glyphicon glyphicon-plus c-add"></span>
    <span className="c-add-span">Add a new criterion.</span></div>
        const addingDiv = <div className="input-group col-md-12"><input
            className="form-control m-input"
            autoFocus={true}
            type="text"
            defaultValue=""
            onChange={(e) => this.setNewDescription(e.target.value)}
            onBlur={() => this.toggleAdd()}
            onKeyDown={(e) => {
                if (e.key === 'Enter') {
                    this.addNewCriterion();
                } else if (e.key === 'Escape') {
                    this.toggleAdd();
                }
            }}
        />
        </div>;
        return this.state.adding ? addingDiv : addDiv;
    }

    private setNewDescription(input: string) {
        this.setState({
            newCriterion: input,
        })
    }

    private async addNewCriterion() {
        const newCriterion = new GradingCriterion();
        newCriterion.setBenchmarkid(this.props.benchmark.getId());
        newCriterion.setDescription(this.state.newCriterion);
        const ans = await this.props.onAdd(newCriterion);
        if (ans) {
            this.state.criteria.push(ans);
        }
        this.setState({
            adding: false,
        })
    }

    private async editCriterion(c: GradingCriterion, input: string): Promise<boolean> {
        c.setDescription(input);
        return this.props.updateCriterion(c);
    }

    private renderCriteriaList(): JSX.Element {
        return <div>
            {this.state.criteria.map((c, i) => <EditCriterion
                key={i}
                criterion={c}
                customScore={this.props.customScore}
                onUpdate={async (newDescription: string) => {
                    const originalDesc = c.getDescription();
                    const ans = await this.editCriterion(c, newDescription);
                    if (!ans) {
                        c.setDescription(originalDesc);
                    }
                }}
                onDelete={() => this.removeCriterion(c)}
            ></EditCriterion>)}
        </div>
    }

    private async removeCriterion(c: GradingCriterion) {
        const ans = await this.props.deleteCriterion(c);
        if (ans) {
            const newList = this.state.criteria;
            newList.splice(this.state.criteria.indexOf(c), 1);
            this.setState({
                criteria: newList,
                editing: false,
            })
        }
    }

    private toggleEdit() {
        this.setState({
            editing: !this.state.editing,
        })
    }

    private toggleAdd() {
        this.setState({
            adding: !this.state.adding,
        })
    }

    private renderTextView(): JSX.Element {
        return <div
            onClick={() => this.toggleEdit()}
    >{this.props.benchmark.getHeading()}{this.removeButton()}</div>
    }

    private renderEditView(): JSX.Element {
        return <div className="input-group col-md-12">
            <input
                className="form-control m-input"
                autoFocus={true}
                type="text"
                defaultValue={this.state.heading}
                onChange={(e) => this.setHeader(e.target.value)}
                onBlur={() => this.toggleEdit()}
                onKeyDown={(e) => {
                    if (e.key === 'Enter') {
                        this.updateHeader();
                    } else if (e.key === 'Escape') {
                        this.toggleEdit();
                    }
                }}
            /></div>
    }

    private setHeader(newHeader: string) {
        this.setState({
            heading: newHeader,
        })
    }

    private updateHeader() {
        this.props.onUpdate(this.state.heading);
        this.setState({
            editing: false,
            heading: this.props.benchmark.getHeading(),
        });
    }

}