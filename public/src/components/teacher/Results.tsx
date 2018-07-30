import * as React from "react";
import { IAssignment, ICourse, IStudentSubmission, IUser, IUserCourseWithUser } from "../../models";

import { DynamicTable, Row, Search, StudentLab } from "../../components";

interface IResultsProp {
    course: ICourse;
    students: IUserCourseWithUser[];
    labs: IAssignment[];
    onApproveClick: (submissionID: number) => void;
}
interface IResultsState {
    assignment?: IStudentSubmission;
    students: IUserCourseWithUser[];
}
class Results extends React.Component<IResultsProp, IResultsState> {
    constructor(props: IResultsProp) {
        super(props);

        const currentStudent = this.props.students.length > 0 ? this.props.students[0] : null;
        if (currentStudent && currentStudent.course.assignments.length > 0 && currentStudent.course.assignments[0]) {
            this.state = {
                // Only using the first student to fetch assignments.
                assignment: currentStudent.course.assignments[0],
                students: this.props.students,
            };
        } else {
            this.state = {
                assignment: undefined,
                students: this.props.students,
            };
        }
    }

    public render() {
        let studentLab: JSX.Element | null = null;
        const currentStudents = this.props.students.length > 0 ? this.props.students : null;
        if (currentStudents
            && this.state.assignment
            && !this.state.assignment.assignment.isgrouplab
        ) {
            studentLab = <StudentLab
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
                <h1>Result: {this.props.course.name}</h1>
                <Row>
                    <div className="col-lg6 col-md-6 col-sm-12">
                        <Search className="input-group"
                            placeholder="Search for students"
                            onChange={(query) => this.handleOnchange(query)}
                        />
                        <DynamicTable header={this.getResultHeader()}
                            data={this.state.students}
                            selector={(item: IUserCourseWithUser) => this.getResultSelector(item)}
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
        let headers: string[] = ["Name", "Slipdays"];
        headers = headers.concat(this.props.labs.filter((e) => !e.isgrouplab).map((e) => e.name));
        return headers;
    }

    private getResultSelector(student: IUserCourseWithUser): Array<string | JSX.Element> {
        const slipdayPlaceholder = "5";
        let selector: Array<string | JSX.Element> = [student.user.name, slipdayPlaceholder];
        selector = selector.concat(student.course.assignments.filter((e, i) => !e.assignment.isgrouplab).map(
            (e, i) => <a className="lab-result-cell"
                onClick={() => this.handleOnclick(e)}
                href="#">
                {e.latest ? (e.latest.score + "%") : "N/A"}</a>));
        return selector;
    }

    private handleOnclick(item: IStudentSubmission): void {
        this.setState({
            assignment: item,
        });
    }

    private handleOnchange(query: string): void {
        query = query.toLowerCase();
        const filteredData: IUserCourseWithUser[] = [];
        this.props.students.forEach((std) => {
            if (std.user.name.toLowerCase().indexOf(query) !== -1
                || std.user.email.toLowerCase().indexOf(query) !== -1
            ) {
                filteredData.push(std);
            }
        });

        this.setState({
            students: filteredData,
        });
    }

}
export { Results };
