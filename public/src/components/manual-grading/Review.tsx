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
    open: boolean;
    addReview: (review: Review) => Promise<Review | null>;
    updateReview: (review: Review) => Promise<boolean>;
    setOpen: () => void;
}

interface ReviewPageState {
    open: boolean;
    benchmarks: GradingBenchmark[];
    feedback: string;
    ready: boolean;
    editing: boolean;
    score: number;
}

export class ReviewPage extends React.Component<ReviewPageProps, ReviewPageState> {

    constructor(props: ReviewPageProps) {
        super(props);
        this.state = {
            open: false,
            benchmarks: [],
            feedback: "",
            ready: false,
            editing: false,
            score: 0,
        }
    }

    public render() {
            console.log("Review: rendering Review, props open: " + this.props.open + ", state open: " + this.state.open);
            const open = this.props.open && this.state.open;
            return <div className="review">
                <h3 className="a-header" onClick={() => {
                    this.props.setOpen();
                    this.toggleOpen();
                }}>{this.props.assignment.getName()}</h3>
                {open ? this.renderInfo() : null}
                {open ?  this.renderBenchmarkList() : null}
                {open ? this.renderFeedback() : null}
                <div className="r-row">{open ? this.graded() : null}{open ? this.saveButton() : null}</div>
                <div className="r-row">{open ? this.showScore() : null}{open ? this.readyButton() : null}</div>
            </div>
    }
    private renderBenchmarkList(): JSX.Element[] {
        console.log("Review: rendering benchmark list, review in props: " + this.props.review);
        const bms: GradingBenchmark[] = this.props.review?.getReviewsList() ?? this.props.assignment.getGradingbenchmarksList();
        return bms.map((bm, i) => <GradeBenchmark
            key={"bm" + i}
            benchmark={bm}
            addComment={(comment: string) => {
                bm.setComment(comment);
                this.setState({
                    benchmarks: bms,
                });
                this.updateReview();
            }}
            onUpdate={(c: GradingCriterion[]) => {
                bm.setCriteriaList(c);
                this.setState({
                    benchmarks: bms,
                });
                this.updateReview();
            }}
        />)
    }

    private renderFeedback(): JSX.Element {
    const feedbackDiv = <div onDoubleClick={() => this.toggleEdit()}>{this.props.review?.getFeedback() ?? "Add a feedback"}</div>;
    const editFeedbackDiv = <div className="input-group">
    <input
        autoFocus={true}
        type="text"
        defaultValue={this.state.feedback}
        onChange={(e) => this.setFeedback(e.target.value)}
        onKeyDown={(e) => {
            if (e.key === 'Enter') {
                this.updateReview();
            } else if (e.key === 'Escape') {
                this.toggleEdit();
            }
        }}
    /></div>;
    return <div className="feedback">
        {this.state.editing ? editFeedbackDiv : feedbackDiv}
    </div>;
    }


    private renderInfo(): JSX.Element {
        console.log("Review: rendering info");
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
                });
                this.updateReview();
            }}
        >Mark as ready</button>
    }

    private setApprovedString(): string {
        return this.props.submission?.approved ? "Approved" : "Not approved";
    }

    private async updateReview() {
        const r: Review = this.props.review ?? this.makeNewReview();
        r.setReady(this.state.ready);
        r.setReviewsList(this.state.benchmarks);
        r.setScore(this.setScore());
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
        });
    }

    private criteriaTotal(): number {
        let counter = 0;
        const bms: GradingBenchmark[] = this.props.assignment.getGradingbenchmarksList();
        bms.forEach((bm) => {
            bm.getCriteriaList().forEach(() => {
                counter++;
            });
        });
        return counter;
    }

    private gradedTotal(): number {
        let counter = 0;
        const bms = this.props.review?.getReviewsList() ?? this.state.benchmarks;
        bms.forEach((r) => {
            r.getCriteriaList().forEach((c) => {
                if (c.getGrade() !== GradingCriterion.Grade.NONE) {
                    counter++;
                }
            });
        });
        return counter;
    }

    private graded(): JSX.Element {
        return <div className="graded-div">Graded: {this.gradedTotal()}/{this.criteriaTotal()}</div>;
    }

    private showScore(): JSX.Element {
        return <div className="score-div">Score: {this.setScore().toFixed()}%</div>;
    }

    private setScore(): number {
        let passed = 0;
        this.state.benchmarks.forEach((bm) => {
            bm.getCriteriaList().forEach((c) => {
                if (c.getGrade() === GradingCriterion.Grade.PASSED) passed++;
            });
        });
        const scoreNow = passed * 100 / this.criteriaTotal();
        return scoreNow;
    }

    private toggleEdit() {
        this.setState({
            feedback: this.props.review?.getFeedback() ?? "",
            editing: !this.state.editing,
        });
    }

    private toggleOpen() {
        this.setState({
            score: this.props.review?.getScore() ?? 0,
            benchmarks: this.props.review?.getReviewsList() ?? [],
            feedback: this.props.review?.getFeedback() ?? "",
            open: !this.state.open,
        });
    }
}
