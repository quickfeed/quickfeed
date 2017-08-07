import * as React from "react";
import { BootstrapButton } from "../../components";
import { ICourse, IOrganization } from "../../models";

import { CourseManager } from "../../managers/CourseManager";

import { bindFunc, RProp } from "../../helper";

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
        return (
            <div>
                <h1>{getTitleText}</h1>
                <form className={this.props.className ? this.props.className : ""}
                    onSubmit={(e) => this.handleFormSubmit(e)}>
                    <div className="form-group">
                        <label className="control-label col-sm-2">Provider:</label>
                        <div className="col-sm-10">
                            {this.renderProviders()}
                        </div>
                    </div>
                    <div className="form-group" id="organisation-container">
                        {this.state.organisations}
                    </div>
                    {this.renderFormControler("Course Name:",
                        "Enter course name",
                        "name",
                        this.state.name,
                        (e) => this.handleInputChange(e))}
                    {this.renderFormControler("Course code:",
                        "Enter course code",
                        "code",
                        this.state.code,
                        (e) => this.handleInputChange(e))}
                    {this.renderFormControler("Course year:",
                        "Enter year",
                        "year",
                        this.state.year,
                        (e) => this.handleInputChange(e))}
                    {this.renderFormControler("Semester:",
                        "Enter semester",
                        "tag",
                        this.state.tag,
                        (e) => this.handleInputChange(e))}

                    <div className="form-group">
                        <div className="col-sm-offset-2 col-sm-10">
                            <BootstrapButton classType="primary" type="submit">
                                {this.props.courseData ? "Update" : "Create"}
                            </BootstrapButton>
                        </div>
                    </div>
                </form>
            </div>
        );
    }

    private renderProviders(): JSX.Element | JSX.Element[] {
        let providers;
        if (this.props.providers.length > 1) {
            providers = this.props.providers.map((provider, index: number) => {
                return <label className="radio-inline" key={index}>
                    <input type="radio"
                        name="provider"
                        value={provider}
                        defaultChecked={this.props.courseData
                            && this.props.courseData.provider === provider ? true : false}
                        onClick={(e) => this.getOrganizations((e.target as HTMLInputElement).value)}
                    />{provider}
                </label>;
            });
        } else {
            const curProvider = this.props.providers[0];
            providers = <label className="radio-inline">
                <input type="hidden"
                    name="provider"
                    value={curProvider}
                />{curProvider}
            </label>;
            if (this.state.provider !== curProvider) {
                this.getOrganizations(curProvider);
            }
        }
        return providers;
    }

    private renderFormControler(
        title: string,
        placeholder: string,
        name: string,
        value: any,
        onChange: (e: React.ChangeEvent<HTMLInputElement>) => void,
    ) {
        return <div className="form-group">
            <label className="control-label col-sm-2" htmlFor="name">{title}</label>
            <div className="col-sm-10">
                <input type="text" className="form-control"
                    id={name}
                    placeholder={placeholder}
                    name={name}
                    value={value}
                    onChange={onChange}
                />
            </div>
        </div>;
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

    private async getOrganizations(provider: string): Promise<void> {
        const directories = await this.props.courseMan.getDirectories(provider);
        this.setState({
            provider,
        });
        this.updateOrganisationDivs(directories);
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
