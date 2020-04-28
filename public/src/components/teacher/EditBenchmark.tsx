import * as React from "react";
import { GradingBenchmark, GradingCriterion } from '../../../proto/ag_pb';
import { EditCriterion } from './EditCriterion';

interface EditBenchmarkProps {
    benchmark: GradingBenchmark,
    onAdd: (c: GradingCriterion) => Promise<GradingCriterion | null>;
    onUpdate: (newHeading: string) => boolean;
    onDelete: () => boolean;

    updateCriterion: (c: GradingCriterion) => boolean;
    deleteCriterion: (id: number) => boolean;
}

interface EditBenchmarkState {
    editing: boolean;
    adding: boolean;
    name: string;
    criteria: GradingCriterion[];
    newCriterion: string;
}

export class EditBenchmark extends React.Component<EditBenchmarkProps, EditBenchmarkState> {

    constructor(props: EditBenchmarkProps) {
        super(props);
        this.state = {
            editing: false,
            adding: false,
            name: this.props.benchmark.getHeading(),
            criteria: this.props.benchmark.getCriteriaList(),
            newCriterion: "",
        }
    }

    public render() {
        return <div>
            <h3 onDoubleClick={() => this.toggleEdit()}>
                {this.state.editing ? this.renderHeader() : this.state.name}
            </h3>

        {this.renderCriteriaList()}

        {this.renderAddRow() }
        </div>
    }

    private renderAddRow(): JSX.Element {
        const addDiv = <div onDoubleClick={() => this.toggleAdd()}>Add a new grading criterion.</div>;
        const addingDiv = <div className="input-btns"><input
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

    private renderCriteriaList(): JSX.Element {
        return <div>
            {this.state.criteria.map(c => <EditCriterion
                key={c.getId()}
                criterion={c}
                onUpdate={(newDescription: string) => {
                    c.setDescription(newDescription);
                    return this.props.updateCriterion(c);
                }}
                onDelete={() => this.props.deleteCriterion(c.getId())}
            ></EditCriterion>)}
        </div>
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
        return <div className="input-btns">
            <input
                type="text"
                defaultValue={this.state.name}
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
            name: newHeader,
        })
    }

    private updateHeader() {
        this.setState({
            editing: false,
        }, () => {
            if (!this.props.onUpdate(this.state.name)) {
                this.setState({
                    name: this.props.benchmark.getHeading(),
                })
            }
        })
    }

}