import * as React from "react";
import { Assignment, GradingBenchmark, GradingCriterion, Review, Submission } from "../../../proto/ag_pb";
import { totalScore } from '../../componentHelper';
import { DynamicTable } from "../data/DynamicTable";
import { ISubmission } from "../../models";

interface ReleaseProps {
    submission: ISubmission | undefined;
    assignment: Assignment;
    authorName: string;
    authorLogin: string;
    studentNumber: number;
    courseURL: string;
    setGrade: (status: Submission.Status) => void;
    release: (ready: boolean) => void;
    getReviewers: (submissionID: number) => Promise<string[]>;
}

interface ReleaseState {
    open: boolean;
    reviews: Review[];
    reviewers: string[];
    score: number;
}
export class Release extends React.Component<ReleaseProps, ReleaseState>{

    constructor(props: ReleaseProps) {
        super(props);
        this.state = {
            reviews: [],
            score: 0,
            reviewers: [],
            open: false,
        }
    }

    public render() {
        const open = this.state.open;
        const reviewInfoSpan = <span className="r-info">Reviews: {this.props.submission?.reviews.length ?? 0}/{this.props.assignment.getReviewers()}</span>;
        const noReviewsSpan = <span className="r-info">N/A</span>;
        const noSubmissionDiv = <div className="alert alert-info">No submissions for {this.props.assignment.getName()}</div>;
        const noReviewsDiv = <div className="alert alert-info">{this.props.assignment.getName()} is not for manual grading</div>
        const noReadyReviewsDiv = <div className="alert alert-info">No ready reviews for {this.props.assignment.getName()}</div>

        const headerDiv = <div className="row review-header" onClick={() => this.toggleOpen()}>
        <h3><span className="r-header">{this.props.studentNumber}. {this.props.authorName}</span><span className="r-score">Score: {this.props.submission?.score ?? 0} </span>{this.props.assignment.getReviewers() > 0 ? reviewInfoSpan : noReviewsSpan}{this.releaseButton()}</h3>
        </div>;


        if (this.props.assignment.getReviewers() < 1) {
            return <div className="release">
                {headerDiv}
                {open ? noReviewsDiv : null}
            </div>
        }

        if (!this.props.submission) {
            return <div className="release">
                {headerDiv}
                {open ? noSubmissionDiv : null}
            </div>
        }

        if (this.state.reviews.length < 1) {
            return <div className="release">
                {headerDiv}
                {open ? noReadyReviewsDiv : null}
            </div>
        }

        return <div className="release">
            {headerDiv}
        ></div>
    }

    private releaseButton(): JSX.Element {
        return <div
            className={this.releaseButtonClass()}
            onClick={() => {
                if (this.props.submission && this.props.assignment.getReviewers() > 0 &&
                this.state.reviews.length === this.props.assignment.getReviewers()) {
                    this.props.release(!this.props.submission.released);
                }
            }}
            >{this.releaseButtonString()}</div>
        }

    private releaseButtonClass(): string {
        if (!this.props.submission || this.props.assignment.getReviewers() < 1 ||
         this.state.reviews.length < this.props.assignment.getReviewers()) {
             return "btn btn-basic disabled release-btn";
         }
        return "btn btn-default release-btn";
    }

    private releaseButtonString(): string {
        if (!this.props.submission || this.props.assignment.getReviewers() < 1 ||
         this.state.reviews.length < this.props.assignment.getReviewers()) {
             return "N/A";
         }
        return this.props.submission.released ? "Released" : "Release"
    }

    private selectReadyReviews(): Review[] {
        const selected: Review[] = [];
        this.props.submission?.reviews.forEach(r => {
            if (r.getReady()) selected.push(r);
        });
        return selected;
    }

    private renderReviewers(): JSX.Element {
        return <ul className="r-list">
            {this.state.reviewers.map((r, i) => <li key={"rl" + i}>
                {r}
            </li>)}
        </ul>;
    }

    private renderReviewTable(): JSX.Element {
        return <div>
            {this.props.assignment.getGradingbenchmarksList().map((bm, i) => <DynamicTable
                header={this.makeHeader(bm)}
                data={bm.getCriteriaList()}
                selector={(c: GradingCriterion) => this.reviewSelector(c)}
            />)}
        </div>
    }

    private reviewSelector(c: GradingCriterion): (string | JSX.Element)[] {
        const tableRow: (string | JSX.Element)[] = [c.getDescription()];
        return tableRow.concat(this.state.reviews.map(rv => this.cellElement(this.chooseCriterion(c.getId(), rv.getBenchmarksList()) ?? c)));
    }

    private makeHeader(bm: GradingBenchmark): (string | JSX.Element)[] {
        const headers: (string | JSX.Element)[] = [bm.getHeading()];
        return headers.concat(this.state.reviews.map(() => <span className="glyphicon glyphicon-comment" onClick={() => this.showBenchmarkComment(bm)}></span>));
    }

    private renderStatusButton(): JSX.Element {
        return <div className="input-group">
            <label className="input-group-addon" htmlFor="submissionStatus">Set status:</label>
            <select className="form-control" id="submissionStatus">
                <option onSelect={() => this.updateStatus(Submission.Status.NONE)}>None</option>
                <option onSelect={() => this.updateStatus(Submission.Status.APPROVED)}>Approved</option>
                <option onSelect={() => this.updateStatus(Submission.Status.REJECTED)}>Rejected</option>
                <option onSelect={() => this.updateStatus(Submission.Status.REVISION)}>Revision</option>
            </select>
            </div>;
    }

    private updateStatus(status: Submission.Status) {
        if (this.props.submission) {
            this.props.setGrade(status);
        }
    }

    private showBenchmarkComment(bm: GradingBenchmark) {
        console.log("Benchmark comment is: " + bm.getComment());
    }

    private showCriterionComment(c: GradingCriterion) {
        console.log("Criterion comment: " + c.getComment());
    }

    private cellElement(c: GradingCriterion): JSX.Element {
        const commentSpan = c.getComment() !== "" ? <span className="glyphicon glyphicon-comment" onClick={() => this.showCriterionComment(c)}/> : "";
        switch (c.getGrade()) {
            case GradingCriterion.Grade.PASSED:
                return <div className="f-cell-pass">
                    <span className="glyphicon glyphicon-plus-sign">{commentSpan}</span>
                </div>;
            case GradingCriterion.Grade.FAILED:
                return <div className="f-cell-fail">
                    <span className="glyphicon glyphicon-minus-sign">{commentSpan}</span>
                </div>;
            default:
                return <div>
                    <span className="glyphicon glyphicon-question-sign">{commentSpan}</span>
                </div>;
        }
    }


    private chooseCriterion(ID: number, bms: GradingBenchmark[]): GradingCriterion | null {
        bms.forEach(bm => {
            const foundC = bm.getCriteriaList().find(c => c.getId() === ID);
            if (foundC) {
                return foundC;
            }
        });
        return null;
    }

    private async toggleOpen() {
        const ready = this.selectReadyReviews();
        if (ready.length > 0) {
            this.setState({
                open: !this.state.open,
                reviewers: this.props.submission ? await this.props.getReviewers(this.props.submission.id) : [],
                reviews: ready,
                score: totalScore(ready),
            });
        } else {
            this.setState({open: !this.state.open});
        }
    }
}