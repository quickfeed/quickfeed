import * as React from "react";
import { Course, GradingBenchmark, GradingCriterion, Assignment, User } from '../../../proto/ag_pb';
import { IStudentLabsForCourse, IReview } from '../../models';


interface GradingViewProps {
    course: Course;
    assignments: Assignment[];
    students: IStudentLabsForCourse[];
    curUser: User;
    addReview: (review: IReview) => Promise<boolean>;
    updateReview: (review: IReview) => Promise<boolean>;

}

interface GradingViewState {
    currentStudent: IStudentLabsForCourse;
}

export class GradingView extends React.Component<GradingViewProps, GradingViewState> {
 // TODO: render a list of all course students to be always displayed as a menu
 // show a corresponding feedbackView for the chosen student to the right of the menu
}