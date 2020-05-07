import * as React from "react";
import { Assignment, GradingBenchmark, GradingCriterion, Review } from '../../../proto/ag_pb';
import { ISubmission } from '../../models';
import { GradeBenchmark } from './GradeBenchmark';

interface ReviewPageProps {
    assignment: Assignment;
    submission: ISubmission | undefined;
    review: Review | null;
    authorName: string;
    reviewerID: number;
    addReview: (review: Review) => Promise<Review | null>;
    updateReview: (review: Review) => Promise<boolean>;
}

interface ReviewPageState {
    open: boolean;
    benchmarks: GradingBenchmark[];
    score: number;
    feedback: string;
    ready: boolean;
}

export class ReviewPage extends React.Component<ReviewPageProps, ReviewPageState> {

    constructor(props: ReviewPageProps) {
        super(props);
        this.state = {
            open: false,
            benchmarks: [],
            score: 0,
            feedback: "",
            ready: false,
        }
    }

    public render() {
        return <div className="review">
            <h3 className="a-header" onClick={() => this.toggleOpen()}>{this.props.assignment.getName()}</h3>
            {this.state.open ? this.renderInfo() : null}
            {this.state.open ?  this.renderBenchmarkList() : null}
            {this.state.open ? this.renderFeedbackRow() : null}
            <div className="r-row">{this.state.open ? this.graded() : null}{this.state.open ? this.saveButton() : null}</div>
            <div className="r-row">{this.state.open ? this.showScore() : null}{this.state.open ? this.readyButton() : null}</div>
        </div>
    }

    private renderBenchmarkList(): JSX.Element[] {
        const bms: GradingBenchmark[] = this.props.review?.getReviewsList() ?? this.props.assignment.getGradingbenchmarksList();
        return bms.map((bm, i) => <GradeBenchmark
            key={"bm" + i}
            benchmark={bm}
            addComment={(comment: string) => {
                bm.setComment(comment);
            }}
            onUpdate={(c: GradingCriterion[]) => {
                bm.setCriteriaList(c);
                this.setState({
                    benchmarks: bms,
                })
            }}
        />)
    }

    private renderFeedbackRow(): JSX.Element {
        return <div className="input-group">
            <div className="input-group-prepend">
        <span className="input-group-text">Add feedback</span>
        </div>
        <textarea
            className="form-control"
            defaultValue={this.state.feedback}
            onChange={(e) => this.setFeedback(e.target.value)}
        ></textarea>
        </div>
    }

    private renderInfo(): JSX.Element {
        return <div className="s-info"><ul>
            <li key="i1"> Reviews: {this.props.submission?.reviews.length ?? 0}/{this.props.assignment.getReviewers()}</li>
            <li key="i2"> {this.setApprovedString()} </li>
            </ul></div>
    }

    private saveButton(): JSX.Element {
        return <button
            onClick={() => {
                this.updateReview();
            }}
        >Save changes</button>
    }

    private readyButton(): JSX.Element {
        return <button
            onClick={() => {
                this.setState({
                    ready: true,
                })
            }}
        >Mark as ready</button>
    }

    private setApprovedString(): string {
        return this.props.submission?.approved ? "Approved" : "Not approved";
    }

    private async updateReview() {
        // TODO: update or add review depending on the context
        const r: Review = this.props.review ?? this.makeNewReview();
        r.setReady(this.state.ready);
        r.setReviewsList(this.state.benchmarks);
        r.setScore(this.state.score);
        r.setFeedback(this.state.feedback);
        if (r.getId() > 0) {
            this.props.updateReview(r);
        } else {
            await this.props.addReview(r);
        }
        this.setScore();
    }

    private makeNewReview(): Review {
        const r = new Review();
        r.setSubmissionid(this.props.submission?.id ?? 0);
        r.setReviewerid(this.props.reviewerID);
        return r;
    }

    private setFeedback(input: string) {
        this.setState({
            feedback: input,
        })
    }

    private criteriaTotal(): number {
        let counter = 0;
        const bms: GradingBenchmark[] = this.props.assignment.getGradingbenchmarksList();
        console.log("Review for " + this.props.authorName + ": got " + bms.length + "benchmarks for this review");
        bms.forEach((bm) => {
            bm.getCriteriaList().forEach(() => {
                counter++;
            });
        });
        console.log("Review: got " + counter + " criterias total for this review");
        return counter;
    }

    private gradedTotal(): number {
        let counter = 0;
        this.state.benchmarks.forEach((r) => {
            r.getCriteriaList().forEach((c) => {
                if (c.getGrade() !== GradingCriterion.Grade.NONE) {
                    counter++;
                }
            });
        });
        return counter;
    }

    private toggleOpen() {
        this.setState({
            open: !this.state.open,
        });
    }

    private graded(): JSX.Element {
        return <div className="graded-div">Graded: {this.gradedTotal()}/{this.criteriaTotal()}</div>;
    }

    private showScore(): JSX.Element {
        return <div className="score-div">Score: {this.props.review?.getScore() ?? this.state.score.toFixed()}%</div>;
    }

    private setScore() {
        let passed = 0;
        this.state.benchmarks.forEach((bm) => {
            bm.getCriteriaList().forEach((c) => {
                if (c.getGrade() === GradingCriterion.Grade.PASSED) passed++;
            })
        });
        const scoreNow = passed * 100 / this.criteriaTotal();
        console.log("Score now: " + scoreNow);
        this.setState({
            score: scoreNow,
        })
    }
}
