import * as React from "react";
import { Assignment, GradingBenchmark, GradingCriterion, Review } from '../../../proto/ag_pb';
import { ISubmission } from "../../models";
import { GradeBenchmark } from "./GradeBenchmark";
import { userSubmissionLink } from "../../componentHelper";

interface ReviewPageProps {
    assignment: Assignment;
    submission: ISubmission | undefined;
    // review: Review | null;
    authorName: string;
    studentNumber: number;
    authorLogin: string;
    courseURL: string;
    reviewerID: number;
    addReview: (review: Review) => Promise<Review | null>;
    updateReview: (review: Review) => Promise<boolean>;
}

interface ReviewPageState {
    review: Review | undefined;
    open: boolean;
    benchmarks: GradingBenchmark[];
    feedback: string;
    ready: boolean;
    editing: boolean;
    score: number;
    alert: string;
    graded: number;
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
            alert: "",
            graded: 0,
            review: undefined,
        }
    }

    public render() {
            const open = this.state.open;
            return <div className="review">
                <div className="row review-header" onClick={() => this.toggleOpen()}>
                <h2><span>{this.props.studentNumber}. {this.props.authorName}</span> <span className="r-info">Reviews: {this.props.submission?.reviews.length ?? 0}/{this.props.assignment.getReviewers()} </span></h2>
                </div>
                {open ? this.renderAlert() : null}

                {open ? this.makeHeaderRow() : null}

                {open ? this.renderInfoTableRow() : null}

                {open ? this.renderBenchmarkList() : null}

                {open ? this.renderFeedback() : null}
            </div>
    }

    private makeHeaderRow(): JSX.Element {
        return <h3>{this.props.submission ? "Submission for " + this.props.assignment.getName() : "No submissions yet for " + this.props.assignment.getName()}</h3>
    }

    private renderBenchmarkList(): JSX.Element {
        const bms: GradingBenchmark[] = this.state.benchmarks;
        return <div className="row">
            {bms.map((bm, i) => <GradeBenchmark
            key={"bm" + i}
            benchmark={bm}
            addComment={(comment: string) => {
                bm.setComment(comment);
                this.updateReview();
            }}
            onUpdate={(c: GradingCriterion[]) => {
                bm.setCriteriaList(c);
                this.updateReview();
            }}
        />)}
        </div>
    }

    private renderFeedback(): JSX.Element {
    const feedbackDiv = <div onDoubleClick={() => this.toggleEdit()}>{"Add a feedback"}</div>;
    const editFeedbackDiv = <div className="input-group col-12">
    <input
        autoFocus={true}
        type="text"
        defaultValue={this.state.review?.getFeedback() ?? this.state.feedback}
        onChange={(e) => this.setFeedback(e.target.value)}
        onKeyDown={(e) => {
            if (e.key === "Enter") {
                this.updateReview();
            } else if (e.key === "Escape") {
                this.toggleEdit();
            }
        }}
    /></div>;
    return <div className="row">
        {this.state.editing ? editFeedbackDiv : feedbackDiv}
    </div>;
    }

    private renderInfoTableRow(): JSX.Element {
        return <div className="row">
            <div className="col-6">
                <ul className="list-group">
                    <li key="li1" className="list-group-item">Score: {this.showScore()}</li>
                    <li key="li2" className="list-group-item">Submission status: {this.props.submission?.status ?? "None"}</li>
                    <li key="li3" className="list-group-item">Review status: {this.state.ready ? "Ready" : "In progress"}</li>
                    <li key="li4" className="list-group-item">Graded: {this.gradedTotal()}/{this.criteriaTotal()}</li>
                </ul>
            </div>
            <div className="col-4">
                <div className="row">
                {this.readyButton()}
                </div>
                <div className="row">
                    {userSubmissionLink(this.props.authorLogin, this.props.assignment.getName(), this.props.courseURL)}
                </div>
            </div>
        </div>;
    }

    private readyButton(): JSX.Element {
        return <button
            onClick={() => {
                if (this.state.review && this.state.review.getReady()) {
                this.setState({
                    ready: false,
                }, () => this.updateReview());
                } else {
                    this.setReady();
                }
            }}
        >Mark as ready</button>
    }

    private setReady() {
        if (this.state.graded < this.criteriaTotal()) {
            this.setAlert("All grading criteria must be checked before marking review as ready.");
        } else {
            this.setState({
                ready: true,
            }, () => this.updateReview());
        }
    }

    private async updateReview(bms?: GradingBenchmark[]) {
        const r: Review = this.state.review ?? this.makeNewReview();
        r.setReady(this.state.ready);
        r.setReviewsList(bms ?? this.state.benchmarks);
        r.setScore(this.setScore());
        r.setFeedback(this.state.feedback);
        if (r.getId() > 0) {
            this.props.updateReview(r);
        } else {
            const rw = await this.props.addReview(r);
            if (rw) {
                this.setState({
                    review: rw,
                });
            }
        }
        this.setState({
            editing: false,
            alert: "",
        });
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
        const bms = this.state.benchmarks;
        bms.forEach((r) => {
            r.getCriteriaList().forEach((c) => {
                if (c.getGrade() !== GradingCriterion.Grade.NONE) {
                    counter++;
                }
            });
        });
        return counter;
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
            feedback: this.state.review?.getFeedback() ?? "",
            editing: !this.state.editing,
        });
    }

    private toggleOpen() {
        const rw = this.selectReview(this.props.submission);
        if (rw) {
            this.setState({
                review: rw,
                score: rw.getScore(),
                benchmarks: this.refreshBenchmarks(rw),
                feedback: rw.getFeedback(),
                open: !this.state.open,
                graded: this.gradedTotal(),
            });
        } else {
            this.setState({
                benchmarks: this.props.assignment.getGradingbenchmarksList(),
                open: !this.state.open,
                graded: this.gradedTotal(),
                score: this.setScore(),
            });
        }
    }

    private renderAlert(): JSX.Element | null {
        return this.state.alert === "" ? null : <div className="row"><div className="alert alert-warning">{ this.state.alert }</div></div>
    }

    private setAlert(alert?: string) {
        this.setState({
            alert: this.makeAlertString(alert),
        });
    }

    private makeAlertString(alert?: string): string {
        if (!this.props.submission) return "";
        if (this.props.assignment.getReviewers() === 0) return "This assignment has no grading criteria";
        if (!this.state.review && this.props.assignment.getReviewers() <= this.props.submission.reviews.length) return "All reviews are redy for this submission";
        return alert ?? "";
    }

    // check and update review in case the assignment benchmarks have been changed
    // after the review had been submitted
    private refreshBenchmarks(r: Review): GradingBenchmark[] {
        const oldList = r.getReviewsList();
        // update benchmarks
        r.getReviewsList().forEach(bm => {
            const assignmentBM = this.props.assignment.getGradingbenchmarksList().find(item => item.getId() === bm.getId());
            // remove deleted benchmarks
            if (!assignmentBM) {
                oldList.splice(oldList.indexOf(bm), 1);
            } else {
                // remove deleted criteria
                const oldCriteriaList = bm.getCriteriaList();
                oldCriteriaList.forEach(c => {
                    if (assignmentBM.getCriteriaList().indexOf(c) < 0) {
                        oldCriteriaList.splice(oldCriteriaList.indexOf(c), 1);
                    }
                });
                // add new criteria
                assignmentBM.getCriteriaList().forEach(c => {
                    if (oldCriteriaList.indexOf(c) < 0) {
                        oldCriteriaList.push(c);
                    }
                });
            }
        });
        // add new benchmarks
        this.props.assignment.getGradingbenchmarksList().forEach(bm => {
            if (oldList.indexOf(bm) < 0) {
                oldList.push(bm);
            }
        });
        return oldList;
    }

    private selectReview(s: ISubmission | undefined): Review | undefined {
        let rw: Review | undefined;
        if (s?.reviews) {
            s.reviews.forEach((r) => {
                if (r.getReviewerid() === this.props.reviewerID) {
                    rw = r;
                }
            });
        }
        return rw;
    }
}
