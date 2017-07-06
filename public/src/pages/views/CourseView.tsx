import * as React from "react";
import { DynamicTable, Search } from "../../components";
import { ICourse } from "../../models";

interface ICourseViewProp {
    courses: ICourse[];
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
        const searchIcon: JSX.Element = <span className="input-group-addon">
            <i className="glyphicon glyphicon-search"></i></span>;
        return (
            <div>
                <Search className="input-group"
                    addonBefore={searchIcon}
                    placeholder="Search for courses"
                    onChange={(query) => this.handleOnchange(query)}
                />
                <DynamicTable
                    header={["ID", "Name", "Tag", "Year/Semester"]}
                    data={this.state.courses}
                    selector={(e: ICourse) => [e.id.toString(), e.name, e.tag, e.year]}
                >
                </DynamicTable>
            </div>
        );
    }

    private handleOnchange(query: string): void {
        query = query.toLowerCase();
        const filteredData: ICourse[] = [];
        this.props.courses.forEach((course) => {
            if (course.name.toLowerCase().indexOf(query) !== -1
                || course.tag.toLowerCase().indexOf(query) !== -1
                || course.year.toLowerCase().indexOf(query) !== -1
            ) {
                filteredData.push(course);
            }
        });

        this.setState({
            courses: filteredData,
        });
    }
}
