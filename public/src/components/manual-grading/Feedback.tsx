import * as React from "react";
import { Assignment, GradingBenchmark, GradingCriterion, Review, User, Submission } from '../../../proto/ag_pb';
import { userSubmissionLink } from '../../componentHelper';
import { DynamicTable } from '../data/DynamicTable';
import { ISubmission } from "../../models";

interface FeedbackProps {
    reviewers: string[];
    submission: ISubmission;
    assignment: Assignment;
    student: User;
    courseURL: string;
    teacherPageView: boolean;
    courseCreatorView: boolean;
    setApproved?: (submissionID: number, status: Submission.Status) => void;
    setReady?: (submissionID: number, ready: boolean) => void;
}

interface FeedbackState {
    reviews: Review[];
}
export class Feedback extends React.Component<FeedbackProps, FeedbackState>{

    constructor(props: FeedbackProps) {
        super(props);
        this.state = {
            reviews: this.props.teacherPageView ? this.props.submission.reviews : this.selectReadyReviews(),
        }
    }

    public render() {
        if (this.state.reviews.length < 1 || (this.props.teacherPageView && !this.props.submission.feedbackReady)) {
            return <div>No ready reviews yet for submission by {this.props.student.getName()}</div>
        }
        if (this.props.courseCreatorView) {
            return <div className="feedback">
            <h3>Reviews for submission for lab {this.props.assignment.getName()} by {this.props.student.getName}}</h3>
            {userSubmissionLink(this.props.student.getLogin(), this.props.assignment.getName(), this.props.courseURL)}
            {this.renderReviewers()}
            {this.renderReviewTable()}
            {this.renderButtons()}
            </div>;
        }
        if (this.props.teacherPageView && this.props.submission.feedbackReady) {
            return <div className="Feedback">
                {this.renderReviewTable()}
            </div>;
        }
        return <div className="feedback">A general view accessible by all teachers and TAs</div>;
        // TODO: show general review info after the lab has been marked as open for feedback by the course teacher
    }

    private selectReadyReviews(): Review[] {
        const selected: Review[] = [];
        this.props.submission.reviews.forEach(r => {
            if (r.getReady()) selected.push(r);
        });
        return selected;
    }

    private renderReviewers(): JSX.Element {
        return <ul className="r-list">
            {this.props.reviewers.map((r, i) => <li key={"rl" + i}>
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
        return tableRow.concat(this.state.reviews.map(rv => this.cellElement(this.chooseCriterion(c.getId(), rv.getReviewsList()) ?? c)));
    }

    private makeHeader(bm: GradingBenchmark): (string | JSX.Element)[] {
        const headers: (string | JSX.Element)[] = [bm.getHeading()];
        return headers.concat(this.state.reviews.map(() => <span className="glyphicon glyphicon-comment" onClick={() => this.showBenchmarkComment(bm)}></span>));
    }

    private renderButtons(): JSX.Element {
        return <div>

        </div>
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

}