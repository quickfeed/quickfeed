import * as React from "react";
import { Assignment, Course } from "../../../proto/ag/ag_pb";
import { DynamicTable } from "../../components";
import { formatDate } from "../../helper";
import { NavigationManager } from "../../managers/NavigationManager";
import { ISubmissionLink } from "../../models";
import { scoreFromReviews } from '../../componentHelper';

interface IPanelProps {
    course: Course;
    labs: ISubmissionLink[];
    navMan: NavigationManager;
}

export class CoursePanel extends React.Component<IPanelProps> {

    public render() {
        const labPath: string = "app/student/courses/" + this.props.course.getId() + "/lab/";
        const glabPath: string = "app/student/courses/" + this.props.course.getId() + "/grouplab/";

        return (
            <div className="col-lg-3 col-md-6 col-sm-6">
                <div className="panel panel-primary">
                    <div className="panel-heading clickable"
                        onClick={() => this.handleCourseClick()}>{this.props.course.getName()}</div>
                    <div className="panel-body">
                        <DynamicTable
                            header={["Labs", "Score", "Deadline"]}
                            data={this.props.labs}
                            selector={(item: ISubmissionLink) => {
                                let score = "N/A";
                                if (item.submission) {
                                    score = item.assignment.getReviewers() > 0 ? scoreFromReviews(item.submission.reviews).toString() : item.submission.score.toString();
                                    score += "%";
                                }
                                return [
                                    item.assignment.getName(),
                                    score,
                                    formatDate(item.assignment.getDeadline()),
                                ];
                            }}
                            onRowClick={(lab: ISubmissionLink) => {
                                const path = !lab.assignment.getIsgrouplab() ? labPath : glabPath;
                                this.handleRowClick(path, lab.assignment);
                            }}
                        />
                    </div>
                </div>
            </div>
        );
    }

    private handleRowClick(pathPrefix: string, lab: Assignment) {
        if (lab) {
            this.props.navMan.navigateTo(pathPrefix + lab.getId());
        }
    }

    private handleCourseClick() {
        const uri: string = "app/student/courses/" + this.props.course.getId();
        this.props.navMan.navigateTo(uri);
    }
}
