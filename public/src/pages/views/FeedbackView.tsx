import * as React from "react";
import { Course, GradingBenchmark, GradingCriterion, Assignment, User } from '../../../proto/ag_pb';
import { IStudentLabsForCourse, IReview, ISubmission, IStudentLab } from '../../models';
import { Review } from '../../components/manual-grading/Review';

interface FeedbackViewProps {
    assignments: Assignment[];
    reviewerID: number;
    student: IStudentLabsForCourse;
    updateReview: (review: IReview) => Promise<boolean>;
}

interface FeedbackViewState {
    name: string,
    submissions: IStudentLab[];
}

export class FeedbackView extends React.Component<FeedbackViewProps, FeedbackViewState> {
    constructor(props: FeedbackViewProps) {
        super(props);
        this.state = {
            name: this.props.student.labs[0]?.authorName ?? "",
            submissions: this.props.student.labs,
        }
    }

    public render() {
        // TODO: a list of all <Feedback> components for all course assignments
        // decide what to show when the assignment is not supposed to be graded manually
        return <div></div>
    }
}