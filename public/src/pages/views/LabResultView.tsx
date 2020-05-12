import * as React from "react";
import { LabResult, LastBuild, LastBuildInfo, Row } from "../../components";
import { IStudentLab, ISubmission } from '../../models';
import { Feedback } from "../../components/manual-grading/Feedback";
import { User, Submission } from '../../../proto/ag_pb';

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
    getReviewers: (submissionID: number) => Promise<string[]>;
    setApproved: (submissionID: number, status: Submission.Status) => void;
    setReady: (submissionID: number, ready: boolean) => void;
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
                            <Row>
                                <div key="loghead" className="col-lg-12">
                                    {this.renderBuildLogOrInfo(buildLog, latest)}
                                </div>
                            </Row>
                        </section>
                    </div>
                </div>
            );
        }
        return <h1>No submissions yet</h1>;
    }

    private async renderBuildLogOrInfo(log: JSX.Element[], latest: ISubmission): Promise<JSX.Element> {
        const buildLog = <div key="logview" className="well"><code id="logs">{log}</code></div>;
        const feedback = <Feedback
        reviewers={await this.props.getReviewers(latest.id)}
        submission={latest}
        assignment={this.props.studentSubmission.assignment}
        student={this.props.student}
        courseURL={this.props.courseURL}
        teacherPageView={this.props.teacherPageView}
        courseCreatorView={this.props.courseCreatorView}
        setApproved={this.props.setApproved}
        setReady={this.props.setReady}
    />
        return this.props.studentSubmission.assignment.getReviewers() > 1 ? feedback : buildLog;
    }
}
