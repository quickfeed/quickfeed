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
    addReview: (review: Review) => Promise<boolean>;
    updateReview: (review: Review) => Promise<boolean>;
}

interface ReviewPageState {
    open: boolean;
    benchmarks: GradingBenchmark[];
    score: number;
    approved: boolean;
    feedback: string;
    editing: boolean;
    criteria: number;
    graded: number;
    ready: boolean;
}

export class ReviewPage extends React.Component<ReviewPageProps, ReviewPageState> {

    constructor(props: ReviewPageProps) {
        super(props);
        this.state = {
            open: false,
            benchmarks: this.setBenchmarks(),
            score: this.props.review?.getScore() ?? 0,
            approved: this.props.submission?.approved ?? false,
            feedback: this.props.review?.getFeedback() ?? "",
            editing: false,
            criteria: this.criteriaTotal(),
            graded: this.gradedTotal(),
            ready: this.props.review?.getReady() ?? false,
        }
    }

    public render() {
        return <div className="review">
            <h3 className="a-header" onClick={() => this.toggleOpen()}>{this.props.assignment.getName()}</h3>
            {this.state.open ? this.renderInfo() : null}
            {this.state.open ?  this.renderBenchmarkList() : null}
            {this.state.open ? this.renderFeedbackRow() : null}
            <div className="row">{this.state.open ? this.graded() : null}{this.state.open ? this.saveButton() : null}</div>
            <div className="row">{this.state.open ? this.showScore() : null}{this.state.open ? this.readyButton() : null}</div>
        </div>
    }

    private renderBenchmarkList(): JSX.Element[] {
        return this.state.benchmarks.map((bm, i) => <GradeBenchmark
            key={"bm" + i}
            benchmark={bm}
            addComment={(comment: string) => {
                bm.setComment(comment);
            }}
            onUpdate={(c: GradingCriterion[]) => {
                bm.setCriteriaList(c);
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
            onKeyDown={(e) => {
                if (e.key === 'Enter') {
                    this.addFeedback();
                }
            }}
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
        const r: Review = this.props.review ?? this.makeNewReview();
        const ans = await this.props.addReview(r);
        if (ans) {
            this.setScore();
        }
    }

    private makeNewReview(): Review {
        const r = new Review();
        r.setSubmissionid(this.props.submission?.id ?? 0);
        r.setReviewerid(this.props.reviewerID);
        r.setReviewsList(this.state.benchmarks);
        r.setReady(this.state.ready);
        r.setFeedback(this.state.feedback);
        r.setScore(this.state.score);
        return r;
    }

    private async addFeedback() {
        const r: Review = this.props.review ?? this.makeNewReview();
        r.setFeedback(this.state.feedback);
        const ans = this.props.updateReview(r);
        if (!ans) {
            this.setState({
                feedback: this.props.review?.getFeedback() ?? "",
            })
        }
    }

    private setFeedback(input: string) {
        this.setState({
            feedback: input,
        })
    }

    private setBenchmarks(): GradingBenchmark[] {
        return this.props.review?.getReviewsList() ?? this.props.assignment.getGradingbenchmarksList();
    }

    private criteriaTotal(): number {
        let counter = 0;
        const bms: GradingBenchmark[] = this.props.review?.getReviewsList() ?? this.props.assignment.getGradingbenchmarksList();
        console.log("criteriaTotal got " + bms.length + "benchmarks");
        bms.forEach((bm) => {
            bm.getCriteriaList().forEach(() => {
                counter++;
            });
        });
        return counter;
    }

    private gradedTotal(): number {
        let counter = 0;
        if (this.props.review) {
            this.props.review.getReviewsList().forEach((r) => {
                r.getCriteriaList().forEach((c) => {
                    if (c.getGrade() !== GradingCriterion.Grade.NONE) {
                        counter++;
                    }
                });
            });
        }
        return counter;
    }

    private toggleOpen() {
        this.setState({
            open: !this.state.open,
        })
    }

    private graded(): JSX.Element {
        return <div className="graded-div">{this.state.graded}/{this.state.criteria}</div>;
    }

    private showScore(): JSX.Element {
        return <div className="score-div">{this.state.score.toFixed()}%</div>;
    }

    private setScore(): number {
        let passed = 0;
        this.state.benchmarks.forEach((bm) => {
            bm.getCriteriaList().forEach((c) => {
                if (c.getGrade() === GradingCriterion.Grade.PASSED) passed++;
            })
        });
        const score = passed * 100 / this.criteriaTotal();
        console.log("Score now: " + score);
        return score;
    }
}
