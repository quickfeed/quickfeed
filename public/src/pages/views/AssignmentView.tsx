import * as React from "react";
import { Assignment, Course, GradingBenchmark, GradingCriterion } from '../../../proto/ag_pb';
import { EditBenchmark } from "../../components/manual-grading/EditBenchmark";

interface AssignmentViewProps {
    assignment: Assignment;
    updateBenchmark: (bm: GradingBenchmark) => Promise<boolean>;
    addBenchmark: (bm: GradingBenchmark) => Promise<GradingBenchmark | null>;
    removeBenchmark: (bm: GradingBenchmark) => Promise<boolean>;
    updateCriterion: (c: GradingCriterion) => Promise<boolean>;
    addCriterion: (c: GradingCriterion) => Promise<GradingCriterion | null>;
    removeCriterion: (c: GradingCriterion) => Promise<boolean>;
}

interface AssignmentViewState {
    adding: boolean;
    open: boolean;
    newBenchmark: string;
    benchmarks: GradingBenchmark[];
}

export class AssigmnentView extends React.Component<AssignmentViewProps, AssignmentViewState> {

    constructor(props: AssignmentViewProps) {
        super(props);
        this.state = {
            adding: false,
            open: false,
            newBenchmark: "",
            benchmarks: this.props.assignment.getGradingbenchmarksList(),
        }
    }

    public render() {
        return <div className="a-element">
            <h3 className="a-header" onClick={() => this.toggleOpen()}>{this.props.assignment.getName()}</h3>
            {this.state.open ? (<div>{this.renderBenchmarks()}</div>) : null}
            {this.state.open ? this.renderAddNew() : null}
        </div>
    }

    private renderBenchmarks(): JSX.Element {
        return <div className="b-list">
            {this.state.benchmarks.map((bm, i) => <EditBenchmark
                key={i}
                benchmark={bm}
                onAdd={(c: GradingCriterion) => {
                    return this.props.addCriterion(c);
                }}
                onUpdate={async (input: string) => {
                    const oldHeading = bm.getHeading();
                    const ans = await this.editBenchmarkHeading(bm, input);
                    if (!ans) {
                        bm.setHeading(oldHeading);
                    }
                }}
                onDelete={() => this.removeBenchmark(bm)}
                updateCriterion={(c: GradingCriterion) => {
                    return this.props.updateCriterion(c);
                }}
                deleteCriterion={(c: GradingCriterion) => this.props.removeCriterion(c)}
            />)}
        </div>
    }

    private async removeBenchmark(bm: GradingBenchmark) {
        const ans = await this.props.removeBenchmark(bm);
        if (ans) {
            const newList = this.state.benchmarks;
            newList.splice(this.state.benchmarks.indexOf(bm), 1)
            this.setState({
                benchmarks: newList,
            })
        }
    }

    private renderAddNew(): JSX.Element {
        const addRow = <div className="add-b" onDoubleClick={() => this.toggleAdding()}>
            Add a new grading benchmark.
        </div>;
        const addingRow = <div className="input-group"><input
        type="text"
        defaultValue=""
        onChange={(e) => this.setNewHeader(e.target.value)}
        />
        <div className="btn-group">
        <button
            className="btn btn-primary btn-xs"
            onClick={() => this.addNewBenchmark()}>OK</button>
        <button
            className="btn btn-danger btn-xs"
            onClick={() => this.toggleAdding()}>X</button></div>
        </div>;
        return this.state.adding ? addingRow : addRow;
    }

    private async editBenchmarkHeading(bm: GradingBenchmark, heading: string): Promise<boolean> {
        bm.setHeading(heading);
        return this.props.updateBenchmark(bm);
    }

    private toggleAdding() {
        this.setState({
            adding: !this.state.adding,
        })
    }

    private setNewHeader(input: string) {
        this.setState({
            newBenchmark: input,
        })
    }

    private async addNewBenchmark() {
        const bm = new GradingBenchmark();
        bm.setHeading(this.state.newBenchmark);
        bm.setAssignmentid(this.props.assignment.getId());
        const ans = await this.props.addBenchmark(bm);
        if (ans) {
            this.state.benchmarks.push(ans);
        }
        this.setState({
            adding: false,
        })
    }

    private toggleOpen() {
        this.setState({
            open: !this.state.open,
        })
    }


}