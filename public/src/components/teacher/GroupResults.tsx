import * as React from "react";
import { Assignment, Course } from "../../../proto/ag_pb";
import { DynamicTable, Row, Search, StudentLab } from "../../components";
import { IStudentLabsForCourse, IStudentLab, ISubmission } from "../../models";
import { ICellElement } from "../data/DynamicTable";
import { generateCellClass, generateGroupRepoLink, sortByScore } from "./labHelper";

interface IResultsProps {
    course: Course;
    courseURL: string;
    groups: IStudentLabsForCourse[];
    labs: Assignment[];
    onApproveClick: (submissionID: number, approved: boolean) => Promise<boolean>;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<ISubmission | null>;
}

interface IResultsState {
    assignment?: IStudentLab;
    groups: IStudentLabsForCourse[];
}

export class GroupResults extends React.Component<IResultsProps, IResultsState> {

    constructor(props: IResultsProps) {
        super(props);

        const currentGroup = this.props.groups.length > 0 ? this.props.groups[0] : null;
        if (currentGroup && currentGroup.labs.length > 0) {
            this.state = {
                // Only using the first group to fetch assignments.
                assignment: currentGroup.labs[0],
                groups: sortByScore(this.props.groups, this.props.labs, true),
            };
        } else {
            this.state = {
                assignment: undefined,
                groups: sortByScore(this.props.groups, this.props.labs, true),
            };
        }
    }

    public render() {
        let groupLab: JSX.Element | null = null;
        const currentGroup = this.props.groups.length > 0 ? this.props.groups : null;
        if (currentGroup
            && this.state.assignment
            && this.state.assignment.assignment.getIsgrouplab()) {
            groupLab = <StudentLab
                assignment={this.state.assignment}
                showApprove={true}
                onRebuildClick={
                    async () => {
                        if (this.state.assignment && this.state.assignment.submission) {
                            const ans = await this.props.onRebuildClick(this.state.assignment.assignment.getId(), this.state.assignment.submission.id);
                            if (ans) {
                                this.state.assignment.submission = ans;
                                return true;
                            }
                        }
                        return false;
                    }
                }
                onApproveClick={(approve: boolean) => {
                    if (this.state.assignment && this.state.assignment.submission) {
                        const ans = this.props.onApproveClick(this.state.assignment.submission.id, approve);
                        if (ans) {
                            this.state.assignment.submission.approved = approve;
                        }
                    }
                }}
            />;
        }

        return (
            <div>
                <h1>Result: {this.props.course.getName()}</h1>
                <Row>
                    <div key="resulthead" className="col-lg6 col-md-6 col-sm-12">
                        <Search className="input-group"
                            placeholder="Search for groups"
                            onChange={(query) => this.handleOnchange(query)}
                        />
                        <DynamicTable header={this.getResultHeader()}
                            data={this.state.groups}
                            selector={(item: IStudentLabsForCourse) => this.getGroupResultSelector(item)}
                        />
                    </div>
                    <div key="resultbody" className="col-lg-6 col-md-6 col-sm-12">
                        {groupLab}
                    </div>
                </Row>
            </div>
        );
    }

    private getResultHeader(): string[] {
        let headers: string[] = ["Name"];
        headers = headers.concat(this.props.labs.filter((e) => e.getIsgrouplab()).map((e) => e.getName()));
        return headers;
    }

    private getGroupResultSelector(group: IStudentLabsForCourse): (string | JSX.Element | ICellElement)[] {
        const grp = group.enrollment.getGroup();
        const name = grp ? generateGroupRepoLink(grp.getName(), this.props.courseURL) : "";
        let selector: (string | JSX.Element | ICellElement)[] = [name];
        selector = selector.concat(group.labs.filter((e, i) => e.assignment.getIsgrouplab()).map(
            (e, i) => {
                let cellCss: string = "";
                if (e.submission) {
                    cellCss = generateCellClass(e);
                }
                const iCell: ICellElement = {
                    value: <a className={cellCss + " lab-cell-link"}
                        onClick={() => this.handleOnclick(e)}
                        href="#">
                        {e.submission ? (e.submission.score + "%") : "N/A"}</a>,
                    className: cellCss,
                };
                return iCell;
            }));
        return selector;
    }

    private handleOnclick(item: IStudentLab): void {
        this.setState({
            assignment: item,
        });
    }

    private handleOnchange(query: string): void {
        query = query.toLowerCase();
        const filteredData: IStudentLabsForCourse[] = [];
        this.props.groups.forEach((std) => {
            const grp = std.enrollment.getGroup();
            const name = grp ? grp.getName() : "";
            if (name.toLowerCase().indexOf(query) !== -1) {
                filteredData.push(std);
            }
        });
        this.setState({
            groups: filteredData,
        });
    }
}
