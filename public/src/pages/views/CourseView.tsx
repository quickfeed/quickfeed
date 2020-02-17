import * as React from "react";
import { Course } from "../../../proto/ag_pb";
import { DynamicTable, Search } from "../../components";

interface ICourseViewProps {
    courses: Course[];
    onEditClick: (id: number) => void;
}

interface ICourseViewState {
    courses: Course[];
}

export class CourseView extends React.Component<ICourseViewProps, ICourseViewState> {
    constructor(props: any) {
        super(props);
        this.state = {
            courses: this.props.courses,
        };
    }

    public render() {

        return (
            <div>
                <Search className="input-group"
                    placeholder="Search for courses"
                    onChange={(query) => this.handleSearch(query)}
                />
                <DynamicTable
                    header={["ID", "Name", "Course Code", "Year", "Semester", "Action"]}
                    data={this.state.courses}
                    selector={(e: Course) =>
                        [e.getId().toString(), e.getName(), e.getCode(), e.getYear().toString(), e.getTag(),
                    <span>
                        <button className="btn btn-primary"
                            onClick={() => this.props.onEditClick(e.getId())}>Edit</button>
                    </span>,
                    ]}
                >
                </DynamicTable>
            </div>
        );
    }

    private handleSearch(query: string): void {
        query = query.toLowerCase();
        const filteredData: Course[] = [];
        this.props.courses.forEach((course) => {
            if (course.getName().toLowerCase().indexOf(query) !== -1
                || course.getCode().toLowerCase().indexOf(query) !== -1
                || course.getYear().toString().indexOf(query) !== -1
            ) {
                filteredData.push(course);
            }
        });
        this.setState({
            courses: filteredData,
        });
    }
}
