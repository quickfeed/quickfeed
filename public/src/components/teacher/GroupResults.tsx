import * as React from "react";
import { Assignment, Course } from "../../../proto/ag_pb";
import { DynamicTable, Row, Search, StudentLab } from "../../components";
import { IAssignmentLink, IStudentSubmission, ISubmission } from "../../models";
import { ICellElement } from "../data/DynamicTable";
import { generateCellClass, generateGroupRepoLink, sortByScore } from "./labHelper";

interface IResultsProps {
    course: Course;
    courseURL: string;
    groups: IAssignmentLink[];
    labs: Assignment[];
    onApproveClick: (submissionID: number, approved: boolean) => Promise<boolean>;
    onRebuildClick: (assignmentID: number, submissionID: number) => Promise<ISubmission | null>;
}

interface IResultsState {
    assignment?: IStudentSubmission;
    groups: IAssignmentLink[];
}

export class GroupResults extends React.Component<IResultsProps, IResultsState> {

    constructor(props: IResultsProps) {
        super(props);

        const currentGroup = this.props.groups.length > 0 ? this.props.groups[0] : null;
        if (currentGroup && currentGroup.assignments.length > 0) {
            this.state = {
                // Only using the first group to fetch assignments.
                assignment: currentGroup.assignments[0],
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
                onApproveClick={(approve: boolean) => {
                    if (this.state.assignment && this.state.assignment.latest) {
                        const ans = this.props.onApproveClick(this.state.assignment.latest.id, approve);
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
                            placeholder="Search for groups"
                            onChange={(query) => this.handleOnchange(query)}
                        />
                        <DynamicTable header={this.getResultHeader()}
                            data={this.state.groups}
                            selector={(item: IAssignmentLink) => this.getGroupResultSelector(item)}
                        />
                    </div>
                    <div className="col-lg-6 col-md-6 col-sm-12">
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

    private getGroupResultSelector(group: IAssignmentLink): Array<string | JSX.Element | ICellElement> {
        const grp = group.link.getGroup();
        const name = grp ? generateGroupRepoLink(grp.getName(), this.props.courseURL) : "";
        let selector: Array<string | JSX.Element | ICellElement> = [name];
        selector = selector.concat(group.assignments.filter((e, i) => e.assignment.getIsgrouplab()).map(
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

    private handleOnclick(item: IStudentSubmission): void {
        this.setState({
            assignment: item,
        });
    }

    private handleOnchange(query: string): void {
        query = query.toLowerCase();
        const filteredData: IAssignmentLink[] = [];
        this.props.groups.forEach((std) => {
            const grp = std.link.getGroup();
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
