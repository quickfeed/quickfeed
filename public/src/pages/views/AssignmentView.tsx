import * as React from "react";
import { Assignment, GradingBenchmark, GradingCriterion } from "../../../proto/ag/ag_pb";
import { EditBenchmark } from "../../components/manual-grading/EditBenchmark";
import { maxAssignmentScore } from "../../componentHelper";

interface AssignmentViewProps {
    assignment: Assignment;
    updateBenchmark: (bm: GradingBenchmark) => Promise<boolean>;
    addBenchmark: (bm: GradingBenchmark) => Promise<GradingBenchmark | null>;
    removeBenchmark: (bm: GradingBenchmark) => Promise<boolean>;
    updateCriterion: (c: GradingCriterion) => Promise<boolean>;
    addCriterion: (c: GradingCriterion) => Promise<GradingCriterion | null>;
    removeCriterion: (c: GradingCriterion) => Promise<boolean>;
    loadBenchmarks: () => Promise<GradingBenchmark[]>;
    rebuildSubmissions: (assignmentID: number, courseID: number) => Promise<boolean>;
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
            benchmarks: [],
            maxScore: this.renderTotalScore(this.props.assignment.getGradingbenchmarksList()),
        }
    }

    public render() {
        const headerDiv = <div className="row"><h3 className="a-header" onClick={() => this.toggleOpen()}>{this.props.assignment.getName()}</h3></div>;
        const noReviewersDiv = <div><div className="alert alert-info">This assignment is not for manual grading</div>{this.testAllButton()}</div>;
        const topDiv = <div className="row top-div"><div className="assignment-p">Reviewers: {this.props.assignment.getReviewers()}</div>
                <div className="score-p">Max points: {this.state.maxScore}</div> {this.loadButton()} </div>;
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
                customScore={this.renderTotalScore() !== 100}
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
                    const newScore = this.state.maxScore - c.getPoints();
                    const ans = await this.props.removeCriterion(c);
                    if (ans) {
                        this.setState({
                            maxScore: newScore,
                        })
                    }
                    return ans;
                }}
            />)}
        </div>
    }

    private renderTotalScore(benchmarks?: GradingBenchmark[]): number {
        if (this.props.assignment.getGradingbenchmarksList().length < 1) {
            return 0;
        }
        const scoreFromCriteria = maxAssignmentScore(benchmarks ?? this.state.benchmarks);
        return scoreFromCriteria > 0 ? scoreFromCriteria : 100;
    }

    private async removeBenchmark(bm: GradingBenchmark) {
        let totalBenchmarkScore = 0;
        bm.getCriteriaList().forEach(c => totalBenchmarkScore += c.getPoints());
        const newScore = this.state.maxScore - totalBenchmarkScore;
        const ans = await this.props.removeBenchmark(bm);
        if (ans) {
            const newList = this.state.benchmarks;
            newList.splice(this.state.benchmarks.indexOf(bm), 1)
            this.setState({
                benchmarks: newList,
                maxScore: newScore,
            });
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
            benchmarks: this.props.assignment.getGradingbenchmarksList(),
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
    private testAllButton(): JSX.Element {
        return <button type="button"
                id="rebuild"
                className="btn btn-default rebuild-button"
            onClick={ () => this.testAll()}
        >Run all tests</button>;
    }

    private async testAll() {
        if (confirm(
            "Warning! This action will run tests for each submission delivered for this assignment. This can take a several minutes."
        )) {
            const ans = await this.props.rebuildSubmissions(this.props.assignment.getId(), this.props.assignment.getCourseid());
            if (ans) {
                // TODO: remove, only for testing
                console.log("Rebuild successful");
            } else {
                console.log("Rebuild failed")
            }
        }
    }

    private testAllButton(): JSX.Element {
        return <button type="button"
                id="rebuild"
                className="btn btn-default rebuild-button"
            onClick={ () => this.testAll()}
        >Run all tests</button>;
    }

    private testAll() {
        if (confirm(
            "Warning! This action will run tests for each submission delivered for this assignment. This can take a several minutes."
        )) {
            this.props.rebuildSubmissions(this.props.assignment.getId(), this.props.assignment.getCourseid());
        }
    }

    private async loadCriteriaFromFile() {
        if (confirm(
            `Warning! This action will replace all assignment grading criteria with criteria from the file.
        All existing reviews for the assignment will also be removed. Proceed?`,
        )) {
            const newBenchmarks = await this.props.loadBenchmarks();
            if (newBenchmarks.length > 0) {
                this.setState({
                    benchmarks: newBenchmarks,
                    maxScore: this.renderTotalScore(newBenchmarks),
                    open: false,
                });
            }
            this.setState({
                open: true,
            })
        }
    }
}