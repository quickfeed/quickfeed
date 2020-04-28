import * as React from "react";
import { Assignment, Course, GradingBenchmark, GradingCriterion } from "../../../proto/ag_pb";
import { EditBenchmark } from "../../components/teacher/EditBenchmark";

interface AssignmentViewProps {
    assignment: Assignment;
    updateBenchmark: (bm: GradingBenchmark) => boolean;
    addBenchmark: (bm: GradingBenchmark) => Promise<GradingBenchmark | null>;
    removeBenchmark: (id: number) => boolean;
    updateCriterion: (c: GradingCriterion) => boolean;
    addCriterion: (c: GradingCriterion) => Promise<GradingCriterion | null>;
    removeCriterion: (criterionID: number, benchmarkID: number) => boolean;
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
            benchmarks: this.props.assignment.getGradingbasisList(),
        }
    }

    public render() {
        return <div>
            <h3 onClick={() => this.toggleOpen()}>{this.props.assignment.getName()}</h3>
            {this.state.open ? (<div>{this.renderBenchmarks()}</div>) : null}
            {this.state.open ? this.renderAddNew() : null}
        </div>
    }

    private renderBenchmarks(): JSX.Element {
        return <div>
            {this.state.benchmarks.map((bm, i) => <EditBenchmark
                key={i}
                benchmark={bm}
                onAdd={(c: GradingCriterion) => {
                    return this.props.addCriterion(c);
                }}
                onUpdate={(input: string) => {
                    bm.setHeading(input);
                    return this.props.updateBenchmark(bm);
                }}
                onDelete={() => this.props.removeBenchmark(bm.getId())}
                updateCriterion={(c: GradingCriterion) => {
                    return this.props.updateCriterion(c);
                }}
                deleteCriterion={(id: number) => this.props.removeCriterion(id, bm.getId())}
            />)}
        </div>
    }

    private renderAddNew(): JSX.Element {
        const addRow = <div onDoubleClick={() => this.toggleAdding()}>
            Add a new grading benchmark.
        </div>;
        const addingRow = <div className="input-btns"><input
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