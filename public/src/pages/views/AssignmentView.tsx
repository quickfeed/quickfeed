import * as React from "react";
import { Assignment, Course, GradingBenchmark, GradingCriterion } from '../../../proto/ag_pb';
import { DynamicTable } from "../../components/data/DynamicTable";
import { BootstrapButton } from "../../components/bootstrap/BootstrapButton";

interface AssignmentViewProps {
    course: Course;
    assignment: Assignment;
    onUpdate: (benchmarkID?: number) => Promise<boolean>
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
            {this.state.open ? (<div>{this.renderBenchmarkTables()}</div>) : null}
        </div>
    }

    private renderBenchmarkTables(): JSX.Element[] {
        const tables: JSX.Element[] = [];
        this.props.benchmarks.forEach((e) => {

            const tab = <DynamicTable
                header={[e.getHeading(), "Action"]}
                data={e.getCriteriaList()}
                selector={(c: GradingCriterion) => [c.getDescription(), this.generateRowButtons(c)]}
                footer={ this.generateFooterRow(e)}
            ></DynamicTable>;
            tables.push(tab);
        });
        return tables;
    }

    private generateRowButtons(c: GradingCriterion): JSX.Element {
        const buttons: JSX.Element[] = [
            <BootstrapButton
            onClick={() => {this.handleEdit(c)}}
            >Edit</BootstrapButton>,
            <BootstrapButton
            classType="danger"
            onClick={() => {this.handleDelete(c)}}
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

    private handleEdit(c: GradingCriterion) {
        console.log("Editing criterion: " + c.getDescription());
    }

    private handleDelete(bm: GradingBenchmark, c: GradingCriterion) {
        console.log("Deleting " + c.getDescription());
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