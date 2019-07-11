import * as React from "react";
import { IAssignment, IGroupCourseWithGroup, IStudentSubmission } from "../../models";

import { Course } from "../../../proto/ag_pb";
import { DynamicTable, Row, Search, StudentLab } from "../../components";

interface IResultsProp {
    course: Course;
    groups: IGroupCourseWithGroup[];
    labs: IAssignment[];
    onApproveClick: (submissionID: number) => void;
}
interface IResultsState {
    assignment?: IStudentSubmission;
    groups: IGroupCourseWithGroup[];
}
class GroupResults extends React.Component<IResultsProp, IResultsState> {
    private approvedStyle = {
        color: "green",
    };

    constructor(props: IResultsProp) {
        super(props);

        const currentGroup = this.props.groups.length > 0 ? this.props.groups[0] : null;
        if (currentGroup && currentGroup.course.assignments.length > 0) {
            this.state = {
                // Only using the first group to fetch assignments.
                assignment: currentGroup.course.assignments[0],
                groups: this.props.groups,
            };
        } else {
            this.state = {
                assignment: undefined,
                groups: this.props.groups,
            };
        }
    }

    public render() {
        let groupLab: JSX.Element | null = null;
        const currentGroup = this.props.groups.length > 0 ? this.props.groups : null;
        if (currentGroup
            && this.state.assignment
            && this.state.assignment.assignment.isgrouplab) {
            groupLab = <StudentLab
                course={this.props.course}
                assignment={this.state.assignment}
                showApprove={true}
                onRebuildClick={() => { }}
                onApproveClick={() => {
                    if (this.state.assignment && this.state.assignment.latest) {
                        this.props.onApproveClick(this.state.assignment.latest.id);
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
                            selector={(item: IGroupCourseWithGroup) => this.getGroupResultSelector(item)}
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
        let headers: string[] = ["Name", "Slipdays"];
        headers = headers.concat(this.props.labs.filter((e) => e.isgrouplab).map((e) => e.name));
        return headers;
    }

    private getGroupResultSelector(group: IGroupCourseWithGroup): Array<string | JSX.Element> {
        const slipdayPlaceholder = "5";
        let selector: Array<string | JSX.Element> = [group.group.getName(), slipdayPlaceholder];
        selector = selector.concat(group.course.assignments.filter((e) => e.assignment.isgrouplab).map((e) => {
            let approvedCss;
            if (e.latest) {
                approvedCss = e.latest.approved ? this.approvedStyle : undefined;
            }
            return <a className="lab-result-cell"
                onClick={() => this.handleOnclick(e)}
                style={approvedCss}
                href="#">
                {e.latest ? (e.latest.score + "%") : "N/A"}</a>;
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
        const filteredData: IGroupCourseWithGroup[] = [];
        this.props.groups.forEach((std) => {
            if (std.group.getName().toLowerCase().indexOf(query) !== -1) {
                filteredData.push(std);
            }
        });

        this.setState({
            groups: filteredData,
        });
    }

}
export { GroupResults };
