import * as React from "react";
import { Assignment, Course, GradingBenchmark, GradingCriterion } from '../../../proto/ag_pb';
import { BootstrapButton } from "../../components/bootstrap/BootstrapButton";
import { EditBenchmark } from "../../components/teacher/EditBenchmark";

interface AssignmentViewProps {
    assignment: Assignment;
    updateBenchmark: (bm: GradingBenchmark) => boolean;
    addBenchmark: (bm: GradingBenchmark) => boolean;
    removeBenchmark: (id: number) => boolean;
    updateCriterion: (c: GradingCriterion) => boolean;
    addCriterion: (c: GradingCriterion) => boolean;
    removeCriterion: (criterionID: number, benchmarkID: number) => boolean;
    benchmarks: GradingBenchmark[];
}

interface AssignmentViewState {
    editing: boolean;
    open: boolean;
}

export class AssigmnentView extends React.Component<AssignmentViewProps, AssignmentViewState> {

    constructor(props: AssignmentViewProps) {
        super(props);
        this.state = {
            editing: false,
            open: false,
        }
    }

    public render() {
        const newBmButton = <BootstrapButton
            onClick = {() => { this.addNewBenchmark("New benchmark header", this.props.assignment.getId())}}
        >Add new grading benchmark</BootstrapButton>
        return <div>
            <h3 onClick={() => this.toggleOpen()}>{this.props.assignment.getName()}</h3>
            {this.state.open ? (<div>{this.renderBenchmarks()}</div>) : null}
        </div>
    }

    private renderBenchmarks(): JSX.Element {
        return <div>
            {this.props.benchmarks.map(bm => <EditBenchmark
                key={bm.getId()}
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

    private generateRowButtons(bm: GradingBenchmark, c: GradingCriterion): JSX.Element {
        const buttons: JSX.Element[] = [
            <BootstrapButton
            onClick={() => {this.handleEdit(c, "New description")}}
            >Edit</BootstrapButton>,
            <BootstrapButton
            classType="danger"
            onClick={(e) => {this.handleDelete(bm, c)}}
            >Delete</BootstrapButton>
        ]
        return <div className="btn-group action-btn">{buttons}</div>;
    }

    private generateFooterRow(bm: GradingBenchmark): JSX.Element[] {
        const btn = <BootstrapButton
        className="btn-benchmark"
        onClick={() => this.addNewCriteria(bm, "new criterion text here") }
        >
        Add new criterion</BootstrapButton>
        return [<div>New criterion text placeholder</div>];
    }

    private handleEdit(c: GradingCriterion, newText: string) {
        c.setDescription(newText);
    }

    private handleDelete(bm: GradingBenchmark, c: GradingCriterion) {
        const list = bm.getCriteriaList();
        list.splice(list.indexOf(c), 1);
        bm.setCriteriaList(list);
        // TODO: try a oneliner here
    }
    private addNewCriteria(bm: GradingBenchmark, description: string) {
        const c = new GradingCriterion();
        c.setDescription(description);
        c.setBenchmarkid(bm.getId());
        bm.addCriteria(c);
        // update server
    }

    private addNewBenchmark(heading: string, assignmentID: number) {
        const bm = new GradingBenchmark();
        bm.setHeading(heading);
        bm.setAssignmentid(assignmentID);
        // update server
    }

    private toggleOpen() {
        this.setState({
            open: !this.state.open,
        })
    }


}