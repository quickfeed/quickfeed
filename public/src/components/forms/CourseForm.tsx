import * as React from "react";
import { Button } from "../../components";
import { ICourse, IOrganization } from "../../models";

import { CourseManager } from "../../managers/CourseManager";

interface ICourseFormProps<T> {
    className?: string;
    courseMan: CourseManager;
    onSubmit: (formData: object, errors: string[]) => void;
    courseData?: ICourse; // for editing an existing course
    providers: string[];
}

interface ICourseFormStates {
    name: string;
    code: string;
    tag: string;
    year: string;
    provider: string;
    directoryid: number;
    organisations: JSX.Element | null;
}

interface ICourseFormData {
    id?: number;
    name: string;
    code: string;
    tag: string;
    year: number;
    provider: string;
    directoryid: number;
}

class CourseForm<T> extends React.Component<ICourseFormProps<T>, ICourseFormStates> {
    constructor(props: any) {
        super(props);
        this.state = {
            name: this.props.courseData ? this.props.courseData.name : "",
            code: this.props.courseData ? this.props.courseData.code : "",
            tag: this.props.courseData ? this.props.courseData.tag : "",
            year: this.props.courseData ? this.props.courseData.year.toString() : "",
            provider: this.props.courseData ? this.props.courseData.provider : "",
            directoryid: this.props.courseData ? this.props.courseData.directoryid : 0,
            organisations: null,
        };
    }

    public render() {
        const getTitleText: string = this.props.courseData ? "Edit Course" : "Create New Course";
        const providers = this.props.providers.map((provider) => {
            return <label className="radio-inline">
                <input type="radio"
                    name="provider"
                    value={provider}
                    defaultChecked={this.props.courseData
                        && this.props.courseData.provider === "github" ? true : false}
                    onClick={(e) => this.getOrganizations(e, this.updateOrganisationDivs)}
                />{provider}
            </label>;
        });
        return (
            <div>
                <h1>{getTitleText}</h1>
                <form className={this.props.className ? this.props.className : ""}
                    onSubmit={(e) => this.handleFormSubmit(e)}>
                    <div className="form-group">
                        <label className="control-label col-sm-2">Provider:</label>
                        <div className="col-sm-10">
                            {providers}
                        </div>
                    </div>
                    <div className="form-group" id="organisation-container">
                        {this.state.organisations}
                    </div>

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
                        <label className="control-label col-sm-2" htmlFor="code">Course Code:</label>
                        <div className="col-sm-10">
                            <input type="text"
                                className="form-control"
                                id="code"
                                placeholder="Enter course code"
                                name="code"
                                value={this.state.code}
                                onChange={(e) => this.handleInputChange(e)}
                            />
                        </div>
                    </div>

                    <div className="form-group">
                        <label className="control-label col-sm-2" htmlFor="year">Year:</label>
                        <div className="col-sm-10">
                            <input type="text"
                                className="form-control"
                                id="year"
                                placeholder="Enter year"
                                name="year"
                                value={this.state.year}
                                onChange={(e) => this.handleInputChange(e)}
                            />
                        </div>
                    </div>
                    <div className="form-group">
                        <label className="control-label col-sm-2" htmlFor="tag">Semester:</label>
                        <div className="col-sm-10">
                            <input type="text"
                                className="form-control"
                                id="tag"
                                placeholder="Enter semester"
                                name="tag"
                                value={this.state.tag}
                                onChange={(e) => this.handleInputChange(e)}
                            />
                        </div>
                    </div>

                    <div className="form-group">
                        <div className="col-sm-offset-2 col-sm-10">
                            <Button className="btn btn-primary"
                                text={this.props.courseData ? "Update" : "Create"}
                                type="submit" />
                        </div>
                    </div>
                </form>
            </div>
        );
    }

    private handleFormSubmit(e: React.FormEvent<any>) {
        e.preventDefault();
        const errors: string[] = this.courseValidate();
        const courseData: ICourseFormData = {
            name: this.state.name,
            code: this.state.code,
            tag: this.state.tag,
            year: parseInt(this.state.year, 10),
            provider: this.state.provider,
            directoryid: this.state.directoryid,
        };
        if (this.props.courseData) {
            courseData.id = this.props.courseData.id;
        }

        this.props.onSubmit(courseData, errors);
    }

    private handleInputChange(e: React.FormEvent<any>) {
        const target: any = e.target;
        const value = target.type === "checkbox" ? target.checked : target.value;
        const name = target.name;

        this.setState({
            [name]: value,
        });
    }

    private handleOrgClick(dirId: number) {
        if (dirId) {
            this.setState({
                directoryid: dirId,
            });
        }
    }

    private getOrganizations(e: any, callback: any): void {
        const pvdr: string = e.target.value;
        this.setState({
            provider: pvdr,
        });

        const pRes = this.props.courseMan.getDirectories(pvdr);
        pRes.then((orgs: IOrganization[]) => {
            callback.call(this, orgs);
        });

    }

    private updateOrganisationDivs(orgs: IOrganization[]): void {
        const organisationDetails: JSX.Element[] = [];
        for (let i: number = 0; i < orgs.length; i++) {
            organisationDetails.push(
                <button key={i} className="btn organisation"
                    onClick={() => this.handleOrgClick(orgs[i].id)}
                    title={orgs[i].path}>

                    <div className="organisationInfo">
                        <img src={orgs[i].avatar}
                            className="img-rounded"
                            width={80}
                            height={80} />
                        <div className="caption">{orgs[i].path}</div>
                    </div>
                    <input type="radio" />
                </button>,
            );

        }

        let orgMsg: JSX.Element;
        if (this.state.provider === "github") {
            orgMsg = <p>Select a GitHub organization for your course.
                (Don't see your organization below? Autograder needs access to your organization.
                Grant access <a href="https://github.com/settings/applications" target="_blank"> here</a>.)
            </p>;
        } else {
            orgMsg = <p>Select a GitLab group.</p>;
        }

        const orgDivs: JSX.Element = <div>
            <label className="control-label col-sm-2">Organization:</label>
            <div className="organisationWrap col-sm-10">
                {orgMsg}

                <div className="btn-group organisationBtnGroup" data-toggle="buttons">
                    {organisationDetails}
                </div>
            </div>
        </div>;

        this.setState({
            organisations: orgDivs,
            directoryid: 0,
        });
    }

    private courseValidate(): string[] {
        const errors: string[] = [];
        if (this.state.name === "") {
            errors.push("Course Name cannot be blank");
        }
        if (this.state.code === "") {
            errors.push("Course Tag cannot be blank.");
        }
        if (this.state.tag === "") {
            errors.push("Semester cannot be blank.");
        }
        if (this.state.provider === "") {
            errors.push("Provider cannot be blank.");
        }
        if (this.state.directoryid === 0) {
            errors.push("Organisation cannot be blank.");
        }
        const year = parseInt(this.state.year, 10);
        if (this.state.year === "") {
            errors.push("Year cannot be blank.");
        } else if (Number.isNaN(year)) {
            errors.push("Year must be a number");
        } else if (year < (new Date()).getFullYear()) {
            errors.push("Year must be greater or equal to current year");
        }
        return errors;
    }

}
export { CourseForm };
