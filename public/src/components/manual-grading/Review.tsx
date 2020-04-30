import * as React from "react";
import { Assignment, GradingBenchmark, GradingCriterion } from '../../../proto/ag_pb';
import { ISubmission, IReview } from '../../models';
import { GradeBenchmark } from './GradeBenchmark';

interface ReviewProps {
    assignment: Assignment;
    submission: ISubmission;
    review: IReview | null;
    authorName: string;
    reviewerID: number;
    addFeedback: (review: IReview) => Promise<boolean>;
    addFeedbackText: (feedback: string) => Promise<boolean>;
    
}

interface ReviewState {
    benchmarks: GradingBenchmark[];
    score: number;
    approved: boolean;
    feedback: string;
    editing: boolean;
    criteria: number;
    graded: number;
    ready: boolean;
}

export class Review extends React.Component<ReviewProps, ReviewState> {

    constructor(props: ReviewProps) {
        super(props);
        this.state = {
            benchmarks: this.setBenchmarks(),
            score: this.props.submission.score,
            approved: this.props.submission.approved,
            feedback: this.props.review?.feedback ?? "",
            editing: false,
            criteria: this.criteriaTotal(),
            graded: this.gradedTotal(),
            ready: this.props.review?.ready ?? false,
        }
    }

    public render() {
        return <div className="review">
            <h3 className="a-header">{this.props.assignment.getName()}: {this.props.authorName}</h3>
            {this.renderBenchmarkList()}
            {this.renderFeedbackRow()}
            {this.graded()}{this.saveButton()}
            {this.showScore()}{this.readyButton()}
        </div>
    }

    private renderBenchmarkList(): JSX.Element[] {
        return this.state.benchmarks.map((bm, i) => <GradeBenchmark
            benchmark={bm}
            addComment={(comment: string) => {
                bm.setComment(comment);
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
            onSubmit={() => this.addFeedback()}
        ></textarea>

        </div>
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

    private updateReview() {
        const r: IReview = this.props.review ?? this.makeNewReview();
        this.props.addFeedback(r);
    }

    private makeNewReview(): IReview {
        return {
            submissionID: this.props.submission.id,
            reviewerID: this.props.reviewerID,
            reviews: this.state.benchmarks,
            ready: this.state.ready,
            feedback: this.state.feedback,
        };
    }

    private async addFeedback() {
        const ans = this.props.addFeedbackText(this.state.feedback);
        if (!ans) {
            this.setState({
                feedback: this.props.review?.feedback ?? "",
            })
        }
    }

    private setFeedback(input: string) {
        this.setState({
            feedback: input,
        })
    }

    private setBenchmarks(): GradingBenchmark[] {
        return this.props.review?.reviews ?? this.props.assignment.getGradingbenchmarksList();
    }

    private criteriaTotal(): number {
        let counter = 0;
        const bms: GradingBenchmark[] = this.props.review?.reviews ?? this.props.assignment.getGradingbenchmarksList();
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
            this.props.review.reviews.forEach((r) => {
                r.getCriteriaList().forEach((c) => {
                    if (c.getGrade() !== GradingCriterion.Grade.NONE) {
                        counter++;
                    }
                });
            });
        }
        return counter;
    }

    private graded(): JSX.Element {
        return <div className="graded-div">{this.state.graded}/{this.state.criteria}</div>;
    }

    private showScore(): JSX.Element {
        return <div className="score-div">{this.state.score}%</div>;
    }

    private setScore(): number {
        // TODO: calculate score from number of graded criteria vs total
        return 123;
    }
}
