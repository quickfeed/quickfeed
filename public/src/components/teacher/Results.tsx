import * as React from "react";
import { Assignment, Course, Submission } from "../../../proto/ag_pb";
import { DynamicTable, Row, Search, StudentLab } from "../../components";
import { IAssignmentLink, IStudentSubmission, ISubmission } from "../../models";
import { ICellElement } from "../data/DynamicTable";
import { generateCellClass, sortByScore } from "./labHelper";

interface IResultsProp {
    course: Course;
    courseURL: string;
    students: IAssignmentLink[];
    labs: Assignment[];
    onApproveClick: (submissionID: number, approve: boolean) => Promise<boolean>;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<ISubmission | null>;
}

interface IResultsState {
    assignment?: IStudentSubmission;
    students: IAssignmentLink[];
}

export class Results extends React.Component<IResultsProp, IResultsState> {

    constructor(props: IResultsProp) {
        super(props);

        const currentStudent = this.props.students.length > 0 ? this.props.students[0] : null;
        const courseAssignments = currentStudent ? currentStudent.course.getAssignmentsList() : null;
        if (currentStudent && courseAssignments && courseAssignments.length > 0) {
            this.state = {
                // Only using the first student to fetch assignments.
                assignment: currentStudent.assignments[0],
                students: sortByScore(this.props.students, this.props.labs, false),
            };
        } else {
            this.state = {
                assignment: undefined,
                students: sortByScore(this.props.students, this.props.labs, false),
            };
        }
    }

    public render() {
        let studentLab: JSX.Element | null = null;
        const currentStudents = this.props.students.length > 0 ? this.props.students : null;
        if (currentStudents
            && this.state.assignment
            && !this.state.assignment.assignment.getIsgrouplab()
        ) {
            studentLab = <StudentLab
                assignment={this.state.assignment}
                showApprove={true}
                onRebuildClick={
                    async () => {
                        if (this.state.assignment && this.state.assignment.latest) {
                            const ans = await this.props.onRebuildClick(this.state.assignment.assignment.getId(), this.state.assignment.latest.id);
                            if (ans) {
                                this.state.assignment.latest = ans;
                                return true;
                            }
                        }
                        return false;
                    }
                }
                onApproveClick={ async (approve: boolean) => {
                    if (this.state.assignment && this.state.assignment.latest) {
                        const ans = await this.props.onApproveClick(this.state.assignment.latest.id, approve);
                        if (ans) {
                            this.state.assignment.latest.approved = approve;
                        }
                    }
                }}
            />;
        }

        return (
            <div>
                <h1>Result: {this.props.course.getName()}</h1>
                <Row>
                    <div className="col-lg6 col-md-6 col-sm-12">
                        <Search className="input-group"
                            placeholder="Search for students"
                            onChange={(query) => this.handleOnchange(query)}
                        />
                        <DynamicTable header={this.getResultHeader()}
                            data={this.state.students}
                            selector={(item: IAssignmentLink) => this.getResultSelector(item)}
                        />
                    </div>
                    <div className="col-lg-6 col-md-6 col-sm-12">
                        {studentLab}
                    </div>
                </Row>
            </div>
        );
    }

    private getResultHeader(): string[] {
        let headers: string[] = ["Name"];
        headers = headers.concat(this.props.labs.filter((e) => !e.getIsgrouplab()).map((e) => e.getName()));
        return headers;
    }

    private getResultSelector(student: IAssignmentLink): Array<string | JSX.Element | ICellElement> {
        const user = student.link.getUser();
        const displayName = user ? this.generateUserRepoLink(user.getName(), user.getLogin()) : "";
        let selector: Array<string | JSX.Element | ICellElement> = [displayName];
        selector = selector.concat(student.assignments.filter((e, i) => !e.assignment.getIsgrouplab()).map(
            (e, i) => {
                let cellCss: string = "";
                if (e.latest) {
                    cellCss = generateCellClass(e);
                }
                const iCell: ICellElement = {
                    value: <a className={cellCss + " lab-cell-link"}
                        onClick={() => this.handleOnclick(e)}
                        href="#">
                        {e.latest ? (e.latest.score + "%") : "N/A"}</a>,
                    className: cellCss,
                };
                return iCell;
            }));
        return selector;
    }

    private generateUserRepoLink(name: string, userName: string): JSX.Element {
        return <a href={this.props.courseURL + userName + "-labs"} target="_blank">{ name }</a>;
    }

    private handleOnclick(item: IStudentSubmission): void {
        this.setState({
            assignment: item,
        });
    }

    private handleOnchange(query: string): void {
        query = query.toLowerCase();
        const filteredData: IAssignmentLink[] = [];
        this.props.students.forEach((std) => {
            const usr = std.link.getUser();
            if (usr) {
                if (usr.getName().toLowerCase().indexOf(query) !== -1
                    || usr.getEmail().toLowerCase().indexOf(query) !== -1
                    || usr.getLogin().toLowerCase().indexOf(query) !== -1
                ) {
                    filteredData.push(std);
                }
            }
        });

        this.setState({
            students: filteredData,
        });
    }
}
