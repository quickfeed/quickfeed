import * as React from "react";
import { Assignment, GradingBenchmark, GradingCriterion, User } from '../../../proto/ag_pb';
import { EditBenchmark } from "../../components/manual-grading/EditBenchmark";
import { ISubmission, IReview } from '../../models';

interface ReviewProps {
    assignment: Assignment;
    submission: ISubmission;
    review: IReview | null;
    authorName: string;
    reviewerID: number;
    addFeedback: (submission: ISubmission) => Promise<boolean>;
}

interface ReviewState {
    score: number;
    approved: boolean;
    feedback: string;
    editing: boolean;
    criteria: number;
    graded: number;
}

export class Review extends React.Component<ReviewProps, ReviewState> {

    constructor(props: ReviewProps) {
        super(props);
        this.state = {
            score: this.props.submission.score,
            approved: this.props.submission.approved,
            feedback: this.props.review?.feedback ?? "",
            editing: false,
            criteria: this.criteriaTotal(),
            graded: this.gradedTotal(),
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

    private criteriaTotal(): number {
        let counter = 0;
        const bms: GradingBenchmark[] = this.props.review?.review ?? this.props.assignment.getGradingbenchmarksList();
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
            this.props.review.review.forEach((r) => {
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
        return <div className="score-div">{this.state.score}</div>;
    }
}
