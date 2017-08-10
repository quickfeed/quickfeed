import * as React from "react";
import { DynamicTable, Search } from "../../components";
import { ICourse } from "../../models";

interface ICourseViewProp {
    courses: ICourse[];
    onEditClick: (id: number) => void;
}

interface ICourseViewState {
    courses: ICourse[];
}

export class CourseView extends React.Component<ICourseViewProp, ICourseViewState> {
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
                    selector={(e: ICourse) => [e.id.toString(), e.name, e.code, e.year.toString(), e.tag,
                    <span>
                        <button className="btn btn-primary"
                            onClick={() => this.props.onEditClick(e.id)}>Edit</button>
                    </span>,
                    ]}
                >
                </DynamicTable>
            </div>
        );
    }

    private handleSearch(query: string): void {
        query = query.toLowerCase();
        const filteredData: ICourse[] = [];
        this.props.courses.forEach((course) => {
            if (course.name.toLowerCase().indexOf(query) !== -1
                || course.code.toLowerCase().indexOf(query) !== -1
                || course.year.toString().indexOf(query) !== -1
            ) {
                filteredData.push(course);
            }
        });

        this.setState({
            courses: filteredData,
        });
    }
}
