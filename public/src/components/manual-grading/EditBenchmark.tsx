import * as React from "react";
import { GradingBenchmark, GradingCriterion } from "../../../proto/ag_pb";
import { EditCriterion } from "./EditCriterion";

interface EditBenchmarkProps {
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
            <h3 className="b-header" onDoubleClick={() => this.toggleEdit()}>
                {this.state.editing ? this.renderHeader() : this.state.heading}{this.removeButton()}
            </h3>

        {this.renderCriteriaList()}

        {this.renderAddRow() }
        </div>
    }

    private removeButton(): JSX.Element {
        return <button className="btn btn-danger btn-xs" onClick={
            () => this.props.onDelete()
        }>X</button>
    }

    private renderAddRow(): JSX.Element {
        const addDiv = <div className="add-b" onDoubleClick={() => this.toggleAdd()}>Add a new grading criterion.</div>;
        const addingDiv = <div className="input-group adding-b"><input
            type="text"
            defaultValue=""
            onChange={(e) => this.setNewDescription(e.target.value)}
        />
        <div className="btn-group">
        <button
            className="btn btn-primary btn-xs"
            onClick={() => this.addNewCriterion()}>OK</button>
        <button
            className="btn btn-danger btn-xs"
            onClick={() => this.toggleAdd()}>X</button></div>
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

    private renderHeader(): JSX.Element {
        return <div className="input-group">
            <input
                type="text"
                defaultValue={this.state.heading}
                onChange={(e) => this.setHeader(e.target.value)}
            />
            <div className="btn-group">
        <button
            className="btn btn-primary btn-xs"
            onClick={() => this.updateHeader()}>OK</button>
        <button
            className="btn btn-danger btn-xs"
            onClick={() => this.toggleEdit()}>X</button></div>
        </div>
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