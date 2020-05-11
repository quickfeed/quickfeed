import * as React from "react";
import { Assignment, GradingBenchmark, GradingCriterion, Review, User } from '../../../proto/ag_pb';
import { ISubmission } from "../../models";
import { GradeBenchmark } from "./GradeBenchmark";
import { userSubmissionLink } from '../../componentHelper';
import { DynamicTable } from '../data/DynamicTable';

interface FeedbackProps {
    reviews: Review[];
    reviewers: User[];
    // submission: ISubmission;
    assignment: Assignment;
    student: User;
    courseURL: string;
    teacherView: boolean;
    setApproved?: () => void;
    setReady?: () => void;
}

export class Feedback extends React.Component<FeedbackProps>{

    public render() {
        if (this.props.reviews.length < 1) {
            return <div>No ready reviews yet for submission by {this.props.student.getName()}</div>
        }
        return <div className="feedback">
            <h3>Reviews for submission for lab {this.props.assignment.getName()} by {this.props.student.getName}}</h3>
            {userSubmissionLink(this.props.student.getLogin(), this.props.assignment.getName(), this.props.courseURL)}
            {this.renderReviewers()}
            {this.renderReviewTable()}
            {this.renderButtons()}
        </div>;
    }

    private renderReviewers(): JSX.Element {
        return <ul className="r-list">
            {this.props.reviewers.map((r, i) => <li key={"rl" + i}>
                {r.getName()}
            </li>)}
        </ul>;
    }

    private renderReviewTable(): JSX.Element {
        return <DynamicTable
            classType={"table-hover table-condensed f-table"}
            header={this.makeHeader()}
            data={this.props.assignment.getGradingbenchmarksList()}
            selector={(item: GradingBenchmark) => this.reviewSelector(item)}
        />
    }

    private reviewSelector(item: GradingBenchmark): (string | JSX.Element)[] {
        return [];
    }

    private makeHeader(): string[] {
        const headers: string[] = ["Criteria"];
        return headers.concat(this.props.reviews.map((r, i) => "#" + (i + 1)));
    }

    private renderButtons(): JSX.Element {
        return <div>

        </div>
    }

}