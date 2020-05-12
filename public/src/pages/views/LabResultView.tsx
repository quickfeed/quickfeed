import * as React from "react";
import { LabResult, LastBuild, LastBuildInfo, Row } from "../../components";
import { IStudentLab } from "../../models";
import { Feedback } from "../../components/manual-grading/Feedback";
import { User } from "../../../proto/ag_pb";

interface ILabInfoProps {
    studentSubmission: IStudentLab;
    student: User;
    courseURL: string;
    showApprove: boolean;
    slipdays: number;
    teacherPageView: boolean;
    courseCreatorView: boolean;
    onApproveClick: (approve: boolean) => void;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<boolean>;
    getReviewers: (submissionID: number) => string[];
    setApproved?: (submissionID: number) => void;
    setReady?: (submissionID: number) => void;
}

export class LabResultView extends React.Component<ILabInfoProps> {

    public render() {
        if (this.props.studentSubmission.submission) {
            const latest = this.props.studentSubmission.submission;
            const buildLog = latest.buildLog.split("\n").map((x, i) => <span key={i} >{x}<br /></span>);
            return (
                <div key="labhead" className="col-md-9 col-sm-9 col-xs-12">
                    <div key="labview" className="result-content" id="resultview">
                        <section id="result">
                            <LabResult
                                assignment_id={this.props.studentSubmission.assignment.getId()}
                                submission_id={latest.id}
                                showApprove={this.props.showApprove}
                                lab={this.props.studentSubmission.assignment.getName()}
                                progress={latest.score}
                                isApproved={latest.approved}
                                authorName={this.props.studentSubmission.authorName}
                                onApproveClick={this.props.onApproveClick}
                                onRebuildClick={this.props.onRebuildClick}
                            />
                            <LastBuildInfo
                                submission={latest}
                                slipdays={this.props.slipdays}
                                assignment={this.props.studentSubmission.assignment}
                            />
                            <LastBuild
                                test_cases={latest.testCases}
                                score={latest.score}
                                scoreLimit={this.props.studentSubmission.assignment.getScorelimit()}
                                weight={100}
                            />
                            <Feedback
                                reviewers={this.props.getReviewers(latest.id)}
                                submission={latest}
                                assignment={this.props.studentSubmission.assignment}
                                student={this.props.student}
                                courseURL={this.props.courseURL}
                                teacherPageView={this.props.teacherPageView}
                                courseCreatorView={this.props.courseCreatorView}
                                setApproved={this.props.setApproved}
                                setReady={this.props.setReady}
                            />
                            <Row>
                                <div key="loghead" className="col-lg-12">
                                    <div key="logview" className="well">
                                        <code id="logs">{buildLog}</code>
                                    </div>
                                </div>
                            </Row>
                        </section>
                    </div>
                </div>
            );
        }
        return <h1>No submissions yet</h1>;
    }

    private renderBuildLogOrInfo(): JSX.Element {
        return <div></div>;
        // TODO: add conditions to render review, build log, or log with feedback
    }
}
