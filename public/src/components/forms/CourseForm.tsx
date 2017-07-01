import * as React from "react";
import {Button} from "../../components";

interface ICourseFormProps<T> {
    className?: string;
    onSubmit: (formData: object, errors: string[]) => void;
}

interface ICourseFormStates {
    name: string;
    tag: string;
    year: string;
}

class CourseForm<T> extends React.Component<ICourseFormProps<T>, ICourseFormStates> {
    constructor(props: any) {
        super(props);
        this.state = {
            name: "",
            tag: "",
            year: "",
        };
    }

    public render() {
        return (
            <form className={this.props.className ? this.props.className : ""}
                  onSubmit={(e) => this.handleFormSubmit(e)}>
                <div className="form-group">
                    <label className="control-label col-sm-2" htmlFor="name">Course Name:</label>
                    <div className="col-sm-10">
                        <input type="text" className="form-control"
                               id="name"
                               placeholder="Enter course name"
                               name="name"
                               value={this.state.name}
                               onChange={(e) => this.handleInputChange(e)}
                        />
                    </div>
                </div>
                <div className="form-group">
                    <label className="control-label col-sm-2" htmlFor="tag">Course Tag:</label>
                    <div className="col-sm-10">
                        <input type="text"
                               className="form-control"
                               id="tag"
                               placeholder="Enter course tag"
                               name="tag"
                               value={this.state.tag}
                               onChange={(e) => this.handleInputChange(e)}
                        />
                    </div>
                </div>
                <div className="form-group">
                    <label className="control-label col-sm-2" htmlFor="tag">Year/Semester:</label>
                    <div className="col-sm-10">
                        <input type="text"
                               className="form-control"
                               id="tag"
                               placeholder="Enter year/semester"
                               name="year"
                               value={this.state.year}
                               onChange={(e) => this.handleInputChange(e)}
                        />
                    </div>
                </div>
                <div className="form-group">
                    <div className="col-sm-offset-2 col-sm-10">
                        <Button className="btn btn-primary" text="Submit" type="submit"/>
                    </div>
                </div>
            </form>
        );
    }

    private handleFormSubmit(e: any) {
        e.preventDefault();
        const errors: string[] = this.courseValidate();
        this.props.onSubmit(this.state, errors);
    }

    private handleInputChange(e: any) {
        const target = e.target;
        const value = target.type === "checkbox" ? target.checked : target.value;
        const name = target.name;

        this.setState({
            [name]: value,
        });
    }

    private courseValidate(): string[] {
        const errors: string[] = [];
        if (this.state.name === "") {
            errors.push("Course Name cannot be blank");
        }
        if (this.state.tag === "") {
            errors.push("Course Tag cannot be blank.");
        }
        if (this.state.year === "") {
            errors.push("Year/Semester cannot be blank.");
        }
        return errors;
    }

}
export {CourseForm};
