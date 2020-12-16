import * as React from "react";
import { Assignment, GradingBenchmark, GradingCriterion, Review, Submission, User } from "../../../proto/ag_pb";
import { scoreFromReviews, userSubmissionLink, setDivider, submissionStatusSelector, getDaysAfterDeadline } from "../../componentHelper";
import { ISubmission } from "../../models";
import { formatDate } from "../../helper";
import ReactTooltip from "react-tooltip";

interface ReleaseProps {

    userIsCourseCreator: boolean;
    submission: ISubmission | undefined;
    assignment: Assignment;
    authorName: string;
    authorLogin: string;
    studentNumber: number;
    courseURL: string;
    teacherView: boolean;
    isSelected: boolean;
    setGrade: (status: Submission.Status, approved: boolean) => Promise<boolean>;
    release: (ready: boolean) => void;
    getReviewers: (submissionID: number) => Promise<User[]>;
}

interface ReleaseState {
    open: boolean;
    reviews: Review[];
    reviewers: Map<User, Review>;
    score: number;
    status: Submission.Status;
}
export class Release extends React.Component<ReleaseProps, ReleaseState>{

    constructor(props: ReleaseProps) {
        super(props);
        this.state = {
            reviews: this.selectReadyReviews(),
            score: scoreFromReviews(this.selectReadyReviews()),
            reviewers: new Map<User, Review>(),
            open: !props.teacherView,
            status: props.submission?.status ?? Submission.Status.NONE,
        }
    }

    public render() {
        const open = this.state.open && this.props.isSelected;
        const reviewInfoSpan = <span className="r-info">Reviews: {this.props.submission?.reviews.length ?? 0}/{this.props.assignment.getReviewers()}</span>;
        const noReviewsSpan = <span className="r-info">N/A</span>;
        const noSubmissionDiv = <div className="alert alert-info">No submissions for {this.props.assignment.getName()}</div>;
        const noReviewsDiv = <div className="alert alert-info">{this.props.assignment.getName()} is not for manual grading</div>
        const noReadyReviewsDiv = <div className="alert alert-info">No ready reviews for {this.props.assignment.getName()}</div>

        const headerDiv = <div className="row review-header" onClick={() => {if (this.props.teacherView) this.toggleOpen()}}>
        <h3><span className="r-number">{this.props.studentNumber}. </span><span className="r-header">{this.props.authorName}</span><span className="r-score">Score: {scoreFromReviews(this.props.submission?.reviews ?? [])} </span>{this.props.assignment.getReviewers() > 0 ? reviewInfoSpan : noReviewsSpan}{this.releaseButton()}</h3>
        </div>;

        if (this.props.assignment.getReviewers() < 1) {
            return <div className="release">
                {headerDiv}
                {open ? noReviewsDiv : null}
            </div>;
        }

        if (!this.props.submission) {
            return <div className="release">
                {headerDiv}
                {open ? noSubmissionDiv : null}
            </div>;
        }

        if (this.selectReadyReviews().length < 1) {
            return <div className="release">
                {headerDiv}
                {open ? noReadyReviewsDiv : null}
            </div>;
        }

        return <div className="release">
            {this.props.teacherView ? headerDiv : null}
            {open ? setDivider() : null}
            {open && this.props.teacherView ? this.infoTable() : null}
            {open ? this.renderReleaseTable() : null}
            {open}
        </div>;
    }

    public componentDidMount() {
        this.mapReviewers();
    }

    private infoTable(): JSX.Element {
        const afterDeadline = this.props.submission ? getDaysAfterDeadline(new Date(this.props.assignment.getDeadline()), this.props.submission.buildDate) : -1;
        return <div className="row">
            <div className="col-md-6 release-info">
                <ul className="list-group">
                    <li key="li0" className="list-group-item r-li">
                        <span className="r-table">Deadline: </span>
                            {formatDate(this.props.assignment.getDeadline())}</li>
                    <li key="li1" className="list-group-item r-li">
                        <span className="r-table">Delivered: </span>
                            {this.props.submission ? formatDate(this.props.submission?.buildDate) + (afterDeadline > 0 ? "   (" + afterDeadline + " days after deadline)" : "") : "Not delivered"}</li>
                    <li key="li3" className="list-group-item r-li">
                        <span className="r-table">Repository: </span>
                        {userSubmissionLink(this.props.authorLogin, this.props.assignment.getName(), this.props.courseURL, "btn btn-default")}</li>
                    <li key="li4" className="list-group-item r-li">{ submissionStatusSelector(this.props.submission?.status ?? 0, (status: string) => this.updateStatus(status), "r-grade")}</li>
                </ul>
            </div>
            <div className="col-md-6">
                <table className="table">
                    <thead><tr key="it">
                            <td key="itd1" >Reviewers:</td>
                            <td key="itd2">Score:</td>
                        </tr></thead>
                        <tbody>
                        {Array.from(this.state.reviewers.keys()).map((r, i) => <tr key={"it" + i}>
                            <td key={"itm " + i}>{r.getName()}</td>
                            <td key={"itr " + i}>{this.state.reviewers.get(r)?.getScore() ?? 0}</td>
                        </tr>)}</tbody>
                </table>
            </div>
        </div>;
    }

    private releaseButton(): JSX.Element {
        return <div
            className={this.releaseButtonClass()}
            onClick={() => {
                if (this.props.submission && this.props.assignment.getReviewers() > 0 && this.props.userIsCourseCreator) {
                    this.props.release(!this.props.submission.released);
                }
            }}>{this.releaseButtonString()}</div>;
        }

    private releaseButtonClass(): string {
        if (!this.props.submission || this.props.assignment.getReviewers() < 1 ||
         this.props.submission.reviews.length < this.props.assignment.getReviewers()) {
             return "btn release-btn";
         }
        return "btn btn-default release-btn";
    }

    private releaseButtonString(): string {
        if (!this.props.submission || this.props.assignment.getReviewers() < 1) {
             return "N/A";
         }
        return this.props.submission.released ? "Released" : "Release";
    }

    private selectReadyReviews(): Review[] {
        const selected: Review[] = [];
        this.props.submission?.reviews.forEach(r => {
            if (r.getReady()) selected.push(r);
        });
        return selected;
    }

    private renderReleaseTable(): JSX.Element {
        const allReviewers = this.props.teacherView ? this.state.reviewers : this.mapReviewersForStudentView();
        const reviewersList = Array.from(allReviewers.keys());
        return <div className="row">
            <table className="table table-condensed table-bordered">
            <thead><tr key="rthead"><th key="th0">Reviews:</th>{reviewersList.map((u, i) => <th key={"th" + (i + 1)} className="release-cell">
                {(allReviewers.get(u)?.getScore() ?? 0) + " %"}
            </th>)}</tr></thead>
            <tbody>
                {this.renderTableRows()}
            </tbody>
            </table>
        </div>;
    }

    private renderTableRows(): JSX.Element[] {
        const rows: JSX.Element[] = [];
        const allReviewers = this.props.teacherView ? this.state.reviewers : this.mapReviewersForStudentView();
        const reviewersList = Array.from(allReviewers.keys());
        this.props.assignment.getGradingbenchmarksList().forEach((bm, i) => {
            rows.push(<tr key={"rt" + i} className="b-header"><td key={"rth" + i}>{bm.getHeading()}</td>{reviewersList.map(u =>
                <td key={"csp" + u.getId()}>{this.commentSpan(this.selectBenchmark(u, bm).getComment(), "bm" + bm.getId())}</td>)}</tr>);
            bm.getCriteriaList().forEach((c, j) => {
                rows.push(<tr key={"rrt" + j + i}><td>{c.getDescription()}</td>
                {reviewersList.map(u => <td key={"rmp" + u.getId()} className={this.setCellColor(u, c)}>
                    <span className={this.setCellIcon(u, c)}></span>
                    {this.commentSpan(this.selectCriterion(u, c).getComment(), "cr" + c.getId())}
                </td>)}
                </tr>);
            });
        });
        rows.push(<tr key="rtf"><td key="fbrow">Feedbacks:</td>
            {reviewersList.map((u, i) => <td key={"fbrow" + i}>{this.commentSpan(allReviewers.get(u)?.getFeedback() ?? "No feedback", "fb" + i)}</td>)}
        </tr>);
        rows.push(<tr key="tscore"><td key="scrow">Score: {this.props.submission?.score ?? 0}</td>
            {reviewersList.map(u => <td key={"scrow" + u.getId()}>{allReviewers.get(u)?.getScore() ?? 0}</td>)}
        </tr>);
        return rows;
    }

    private setCellIcon(u: User, c: GradingCriterion): string {
        const cr = this.selectCriterion(u, c);
        switch (cr.getGrade()) {
            case GradingCriterion.Grade.PASSED:
                return "r-cell glyphicon glyphicon-ok";
            case GradingCriterion.Grade.FAILED:
                return "r-cell glyphicon glyphicon-remove";
            default:
                return "r-cell glyphicon glyphicon-ban-circle";
        }
    }

    private setCellColor(u: User, c: GradingCriterion): string {
            const cr = this.selectCriterion(u, c);
            if (cr.getGrade() === GradingCriterion.Grade.PASSED) {
                return "success";
            }
            return cr.getGrade() === GradingCriterion.Grade.FAILED ? "danger" : "";
    }

    private selectBenchmark(u: User, bm: GradingBenchmark): GradingBenchmark {
        const allReviewers = this.state.reviewers;
        const allReviews = Array.from(allReviewers.values());

        const r = this.props.teacherView ? allReviewers.get(u) : allReviews.find(item => item.getReviewerid() === u.getId());
        if (r) {
            const rbm = r.getBenchmarksList().find(item => item.getId() === bm.getId())
            if (rbm) bm = rbm;
        }
        return bm;
    }

    private selectCriterion(u: User, c: GradingCriterion): GradingCriterion {
        const allReviewers = this.props.teacherView ? this.state.reviewers : this.mapReviewersForStudentView();
        const allReviews = Array.from(allReviewers.values());
        const r = this.props.teacherView ? allReviewers.get(u) : allReviews.find(item => item.getReviewerid() === u.getId());
        if (r) {
            r.getBenchmarksList().forEach(bm => {
                const rc = bm.getCriteriaList().find(item => item.getId() === c.getId());
                if (rc) {
                    c = rc;
                }
            });
        }
        return c;
    }

    private commentSpan(text: string, id: string): JSX.Element {
        if (text === "") {
            return <span></span>;
        }
        return <span><span className="release-comment glyphicon glyphicon-comment"
            data-tip
            data-for={id}
        ></span>
        <ReactTooltip
            type="light"
            effect="solid"
            id={id}
        ><p>{text}</p></ReactTooltip></span>;
    }

    private async updateStatus(action: string) {
        if (this.props.submission) {
            let newStatus: Submission.Status = Submission.Status.NONE;
            let newBool = false;
            switch (action) {
                case "1":
                    newStatus = Submission.Status.APPROVED;
                    newBool = true;
                    break;
                case "2":
                    newStatus = Submission.Status.REJECTED;
                    break;
                case "3":
                    newStatus = Submission.Status.REVISION;
                    break;
                default:
                    newStatus = Submission.Status.NONE;
                    break;
            }
            const ans = await this.props.setGrade(newStatus, newBool);
            if (ans) {
                this.setState({
                    status: newStatus,
                });
            }
        }
    }

    private mapReviewersForStudentView(): Map<User, Review> {
        const reviews = this.selectReadyReviews();
        const newMap = new Map<User, Review>();
        if (this.props.submission) {
            reviews.forEach(rw => {
                const usr = new User();
                usr.setId(rw.getReviewerid());
                newMap.set(usr, rw);
            });
        }
        return newMap;
    }

    private async mapReviewers() {
        const reviews = this.selectReadyReviews();
        const updatedMap = new Map<User, Review>();
        if (this.props.submission && reviews.length > 0) {
            const reviewers = await this.props.getReviewers(this.props.submission.id);
            reviewers.forEach(r => {
                const selectedReview = this.selectReviewByReviewer(r, reviews);
                if (selectedReview) updatedMap.set(r, selectedReview);
            });
        }
        this.setState({
            reviewers: updatedMap,
        });
    }

    private selectReviewByReviewer(user: User, reviews: Review[]): Review | undefined {
        return reviews.find(item => item.getReviewerid() === user.getId());
    }

    private async toggleOpen() {
        // if closing, flush the state
        if (this.state.open) {
            this.setState({
                reviews: [],
                // reviewers: new Map<User, Review>(),
                open: false,
            });
            return;
        }

        const ready = this.selectReadyReviews();
        this.mapReviewers();
        if (ready.length > 0) {
            this.setState({
                open: this.props.isSelected ? !this.state.open : true,
                reviews: ready,
                score: scoreFromReviews(ready),
                status: this.props.submission?.status ?? Submission.Status.NONE,
            });
        } else {
            this.setState({open: !this.state.open});
        }
    }

}