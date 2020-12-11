import * as React from "react";
import { Assignment, GradingBenchmark, GradingCriterion, Benchmarks } from '../../../proto/ag_pb';
import { EditBenchmark } from "../../components/manual-grading/EditBenchmark";
import { maxAssignmentScore } from '../../componentHelper';

interface AssignmentViewProps {
    assignment: Assignment;
    updateBenchmark: (bm: GradingBenchmark) => Promise<boolean>;
    addBenchmark: (bm: GradingBenchmark) => Promise<GradingBenchmark | null>;
    removeBenchmark: (bm: GradingBenchmark) => Promise<boolean>;
    updateCriterion: (c: GradingCriterion) => Promise<boolean>;
    addCriterion: (c: GradingCriterion) => Promise<GradingCriterion | null>;
    removeCriterion: (c: GradingCriterion) => Promise<boolean>;
    loadBenchmarks: () => Promise<GradingBenchmark[]>;
}

interface AssignmentViewState {
    adding: boolean;
    open: boolean;
    newBenchmark: string;
    benchmarks: GradingBenchmark[];
    maxScore: number;
}

export class AssignmentView extends React.Component<AssignmentViewProps, AssignmentViewState> {

    constructor(props: AssignmentViewProps) {
        super(props);
        this.state = {
            adding: false,
            open: false,
            newBenchmark: "",
            benchmarks: this.props.assignment.getGradingbenchmarksList(),
            maxScore: this.renderTotalScore(),
        }
    }

    public render() {
        const headerDiv = <div className="row"><h3 className="a-header" onClick={() => this.toggleOpen()}>{this.props.assignment.getName()}</h3></div>;
        const noReviewersDiv = <div className="alert alert-info">This assignment is not for manual grading</div>;
        const topDiv = <div className="row top-div"><div className="assignment-p">Reviewers: {this.props.assignment.getReviewers()}</div>
                <div className="score-p">Max score: {this.state.maxScore}</div> {this.loadButton()} </div>;
        if (this.props.assignment.getReviewers() < 1) {
            return <div className="a-element">
                {headerDiv}
                {this.state.open ? noReviewersDiv : null}
            </div>
        }
        return <div className="a-element">
            {headerDiv}
            {this.state.open ? topDiv : null}
            {this.state.open ? (<div className="row">{this.renderBenchmarks()}</div>) : null}
            {this.state.open ? this.renderAddNew() : null}
        </div>
    }

    private renderBenchmarks(): JSX.Element {
        return <div className="b-list">
            {this.state.benchmarks.map((bm, i) => <EditBenchmark
                key={i}
                benchmark={bm}
                customScore={this.state.maxScore !== 100}
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
                deleteCriterion={async (c: GradingCriterion) => {
                    const ans = await this.props.removeCriterion(c);
                    if (ans) {
                        this.setState({
                            maxScore: this.renderTotalScore(),
                        });
                    }
                    return ans;
                }}
            />)}
        </div>
    }

    private renderTotalScore(): number {
        if (this.props.assignment.getGradingbenchmarksList().length < 1) {
            return 0;
        }
        return maxAssignmentScore(this.props.assignment.getGradingbenchmarksList() ?? 100);
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
        const addRow =  <div className="row bm-add-row" onClick={() => this.toggleAdd()}>
            <span className="glyphicon glyphicon-plus bm-add"></span>
            <span className="c-add-span"> Add a new grading criteria group.</span></div>
        const addingRow = <div className="input-group col-md-12"><input
        className="form-control m-input"
        autoFocus={true}
        type="text"
        defaultValue=""
        onChange={(e) => this.setNewHeader(e.target.value)}
        onBlur={() => this.toggleAdd()}
        onKeyDown={(e) => {
            if (e.key === "Enter") {
                this.addNewBenchmark();
            } else if (e.key === "Escape") {
                this.toggleAdd();
            }
        }}
        />
        </div>;
        return this.state.adding ? addingRow : addRow;
    }

    private async editBenchmarkHeading(bm: GradingBenchmark, heading: string): Promise<boolean> {
        bm.setHeading(heading);
        return this.props.updateBenchmark(bm);
    }

    private toggleAdd() {
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

    private loadButton(): JSX.Element {
        return <button type="button"
                id="load"
                className="btn btn-default load-button"
                onClick={() => this.loadCriteriaFromFile()}
        >Load from file</button>;
    }

    private async loadCriteriaFromFile() {
        if (confirm(
            `Warning! This action will remove existing criteria and replace them with criteria from the file. Proceed?`,
        )) {
            const newBenchmarks = await this.props.loadBenchmarks();
            if (newBenchmarks.length > 0) {
                this.setState({
                    benchmarks: newBenchmarks,
                    maxScore: maxAssignmentScore(newBenchmarks),
                    open: false,
                });
            }
            this.setState({
                open: true,
            })
        }
    }
}