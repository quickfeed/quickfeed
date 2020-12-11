import * as React from "react";
import { Assignment, GradingBenchmark, GradingCriterion, Review } from "../../../proto/ag_pb";
import { ISubmission } from "../../models";
import { GradeBenchmark } from "./GradeBenchmark";
import { deepCopy, userSubmissionLink, submissionStatusToString, setDivider, maxAssignmentScore } from '../../componentHelper';

interface ReviewPageProps {
    assignment: Assignment;
    submission: ISubmission | undefined;
    authorName: string;
    studentNumber: number;
    authorLogin: string;
    courseURL: string;
    reviewerID: number;
    isSelected: boolean;
    addReview: (review: Review) => Promise<boolean>;
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
    scoreFromCriteria: number;
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
            scoreFromCriteria: maxAssignmentScore(props.assignment.getGradingbenchmarksList()),
        }
    }

    public render() {
        const open = this.state.open && this.props.isSelected;
        const reviewInfoSpan = <span className="r-info">Reviews: {this.props.submission?.reviews.length ?? 0}/{this.props.assignment.getReviewers()}</span>;
        const noReviewsSpan = <span className="r-info">N/A</span>;
        const headerDiv = <div className="row review-header" onClick={() => this.toggleOpen()}>
        <h3><span className="r-number">{this.props.studentNumber}. </span> <span className="r-header">{this.props.authorName}</span>{this.props.assignment.getReviewers() > 0 ? reviewInfoSpan : noReviewsSpan}</h3>
        </div>;

        const noSubmissionDiv = <div className="alert alert-info">No submissions for {this.props.assignment.getName()}</div>;
        const noReviewsDiv = <div className="alert alert-info">{this.props.assignment.getName()} is not for manual grading</div>

        if (this.props.assignment.getReviewers() < 1) {
            return <div className="review">
                {headerDiv}
                {open ? noReviewsDiv : null}
            </div>
        }


        if (!this.props.submission) {
            return <div className="review">
                {headerDiv}
                {open ? noSubmissionDiv : null}
            </div>
        }

        return <div className="review">
            {headerDiv}

            {open ? setDivider() : null}

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
    const feedbackDiv = <div className="row f-row" onClick={() => this.toggleEdit()}>{"Add a summary feedback"}</div>;
    const editFeedbackDiv = <div className="f-row input-group col-md-12">
    <input
        className="form-control m-input"
        autoFocus={true}
        type="text"
        defaultValue={this.state.review?.getFeedback() ?? this.state.feedback}
        onChange={(e) => this.setFeedback(e.target.value)}
        onBlur={() => this.toggleEdit()}
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
            <div className="col-md-10">
                <ul className="list-group">
                    <li key="li1" className="list-group-item r-li"><span className="r-table">Score: </span>{this.scoreString()}</li>
                    <li key="li2" className="list-group-item r-li"><span className="r-table">Submission status: </span>{submissionStatusToString(this.props.submission?.status)}</li>
                    <li key="li3" className="list-group-item r-li"><span className="r-table">Review status: </span>{this.state.ready ? "Ready" : "In progress"}</li>
                    <li key="li4" className="list-group-item r-li"><span className="r-table">Graded: </span>{this.gradedTotal()}/{this.criteriaTotal()}</li>
                </ul>
            </div>
            <div className="col-md-2">
                <div className="row">
                {this.readyButton()}
                </div>
                <div className="row">
                    {userSubmissionLink(this.props.authorLogin, this.props.assignment.getName(), this.props.courseURL, "btn btn-default")}
                </div>
            </div>
        </div>;
    }

    private readyButton(): JSX.Element {
        return <div className="btn btn-default r-btn"
            onClick={() => {
                if (this.state.review && this.state.review.getReady()) {
                this.setState({
                    ready: false,
                }, () => this.updateReview());
                } else {
                    this.setReady();
                }
            }}
        >Mark as ready</div>
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

    private async updateReview() {
        const r: Review = this.state.review ?? this.makeNewReview();
        r.setReady(this.state.ready);
        r.setBenchmarksList(this.state.benchmarks);
        r.setScore(this.setScore());
        r.setFeedback(this.state.feedback);
        if (r.getId() > 0) {
            const ans = await this.props.updateReview(r);
            if (ans) {
                const newRw = this.selectReview(this.props.submission);
                this.setState({
                    review: newRw,
                    benchmarks: newRw?.getBenchmarksList() ?? this.state.benchmarks,
                    graded: this.gradedTotal(newRw),
                });
            }
        } else {
            const ans = await this.props.addReview(r);
            if (ans) {
                const newRw = this.selectReview(this.props.submission);
                this.setState({
                    review: newRw,
                    benchmarks: newRw?.getBenchmarksList() ?? this.state.benchmarks,
                    graded: this.gradedTotal(newRw),

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

    private gradedTotal(rw?: Review): number {
        let counter = 0;
        const bms = rw?.getBenchmarksList() ?? this.state.benchmarks;
        bms.forEach((r) => {
            r.getCriteriaList().forEach((c) => {
                if (c.getGrade() !== GradingCriterion.Grade.NONE) {
                    counter++;
                }
            });
        });
        return counter;
    }

    private scoreString(): string {
        return this.setScore().toPrecision() + " %";
    }

    private setScore(): number {
        return this.state.scoreFromCriteria > 0 ? this.setCustomScore() : this.setFullScore();
    }

    private setFullScore(): number {
        let passed = 0;
        this.state.benchmarks.forEach((bm) => {
            bm.getCriteriaList().forEach((c) => {
                if (c.getGrade() === GradingCriterion.Grade.PASSED) passed++;
            });
        });
        const total = this.criteriaTotal() > 0 ? this.criteriaTotal() : 1;
        const scoreNow = passed * 100 / total;
        return Math.floor(scoreNow);
    }

    private setCustomScore(): number {
        let scoreNow = 0;
        this.state.benchmarks.forEach((bm) => {
            bm.getCriteriaList().forEach((c) => {
                if (c.getGrade() === GradingCriterion.Grade.PASSED) {
                    scoreNow += c.getScore();
                }
            });
        });
        return scoreNow;
    }

    private toggleEdit() {
        this.setState({
            feedback: this.state.review?.getFeedback() ?? "",
            editing: !this.state.editing,
        });
    }

    private toggleOpen() {
        // reset state when closing
        if (this.state.open) {
            this.setState({
                review: undefined,
                score: this.setScore(),
                benchmarks: [],
                feedback: "",
                open: false,
                graded: 0,
                ready: false,
            });
            return;
        }
        const rw = this.selectReview(this.props.submission);
        if (rw) {
            this.setState({
                review: rw,
                score: rw.getScore(),
                benchmarks: this.refreshBenchmarks(rw),
                feedback: rw.getFeedback(),
                open: this.props.isSelected ? !this.state.open : true,
                graded: this.gradedTotal(rw),
                ready: rw.getReady(),

            });
        } else {
            this.setState({
                review: undefined,
                benchmarks: deepCopy(this.props.assignment.getGradingbenchmarksList()),
                open: this.props.isSelected ? !this.state.open : true,
                graded: this.gradedTotal(),
                score: 0,
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
        if (!this.state.review && this.props.assignment.getReviewers() <= this.props.submission.reviews.length) return "All reviews are ready for this submission";
        return alert ?? "";
    }

    // check and update review in case the assignment benchmarks have been changed
    // after the review had been submitted
    private refreshBenchmarks(r: Review): GradingBenchmark[] {
        const oldList = r.getBenchmarksList();
        // update benchmarks
        oldList.forEach(bm => {
            const assignmentBM = this.props.assignment.getGradingbenchmarksList().find(item => item.getId() === bm.getId());
            // remove deleted benchmarks
            if (!assignmentBM) {
                oldList.splice(oldList.indexOf(bm), 1);
            } else {
                // update description in case there were some changes
                bm.setHeading(assignmentBM.getHeading());
                // remove deleted criteria
                const oldCriteriaList = bm.getCriteriaList();
                oldCriteriaList.forEach(c => {
                    const assignmentCriterium = assignmentBM.getCriteriaList().find(item => item.getId() === c.getId());
                    if (!assignmentCriterium) {
                        oldCriteriaList.splice(oldCriteriaList.indexOf(c), 1);
                    } else {
                        c.setDescription(assignmentCriterium.getDescription());
                    }
                });
                // add new criteria
                assignmentBM.getCriteriaList().forEach(c => {
                    if (!oldCriteriaList.find(item => item.getId() === c.getId())) {
                        oldCriteriaList.push(c);
                    }
                });
            }
        });
        // add new benchmarks
        this.props.assignment.getGradingbenchmarksList().forEach(bm => {
            if (!oldList.find(item => item.getId() === bm.getId())) {
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
