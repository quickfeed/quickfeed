import * as React from "react";
import {Button} from "../../components";
import {IOrganization} from "../../models";

import {CourseManager} from "../../managers/CourseManager";

interface ICourseFormProps<T> {
    className?: string;
    courseMan: CourseManager;
    onSubmit: (formData: object, errors: string[]) => void;
}

interface ICourseFormStates {
    name: string;
    tag: string;
    semester: string;
    year: string;
    provider: string;
    directoryid: number;
    organisations: JSX.Element | null;
}

class CourseForm<T> extends React.Component<ICourseFormProps<T>, ICourseFormStates> {
    constructor(props: any) {
        super(props);
        this.state = {
            name: "",
            tag: "",
            semester: "",
            year: "",
            provider: "",
            directoryid: 0,
            organisations: null,
        };
    }

    public render() {
        return (
            <form className={this.props.className ? this.props.className : ""}
                  onSubmit={(e) => this.handleFormSubmit(e)}>
                <div className="form-group">
                    <label className="control-label col-sm-2">Provider:</label>
                    <div className="col-sm-10">
                        <label className="radio-inline">
                            <input type="radio"
                                   name="provider"
                                   value="github"
                                   onClick={(e) => this.getOrganizations(e, this.updateOrganisationDivs)}
                            />Github
                        </label>
                        <label className="radio-inline">
                            <input type="radio"
                                   name="provider"
                                   value="gitlab"
                                   onClick={(e) => this.getOrganizations(e, this.updateOrganisationDivs)}
                            />Gitlab
                        </label>
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
                    <label className="control-label col-sm-2" htmlFor="semester">Semester:</label>
                    <div className="col-sm-10">
                        <input type="text"
                               className="form-control"
                               id="semester"
                               placeholder="Enter semester"
                               name="semester"
                               value={this.state.semester}
                               onChange={(e) => this.handleInputChange(e)}
                        />
                    </div>
                </div>

                <div className="form-group">
                    <div className="col-sm-offset-2 col-sm-10">
                        <Button className="btn btn-primary" text="Create" type="submit"/>
                    </div>
                </div>
            </form>
        );
    }

    private handleFormSubmit(e: React.FormEvent<any>) {
        e.preventDefault();
        const errors: string[] = this.courseValidate();
        const courseData = {
            name: this.state.name,
            tag: this.state.tag,
            semester: this.state.semester,
            year: parseInt(this.state.year, 10),
            provider: this.state.provider,
            directoryid: this.state.directoryid,
        };

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

    private handleOrgClick(e: any) {
        const target = e.target;
        if (target.hasAttribute("data-directoryid")) {
            this.setState({
                directoryid: target.getAttribute("data-directoryid"),
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
                        data-directoryid={orgs[i].id}
                        onClick={(e) => this.handleOrgClick(e)}>

                    <div className="organisationInfo">
                        <img src={orgs[i].avatar}
                             className="img-rounded"
                             width={80}
                             height={80}/>
                        <div className="caption">{orgs[i].path}</div>
                    </div>
                    <input type="radio"/>
                </button>,
            );

        }

        const orgDivs: JSX.Element = <div>
            <label className="control-label col-sm-2">Organisation:</label>
            <div className="organisationWrap col-sm-10">
                <p> Select an Organisation</p>
                <div className="btn-group organisationBtnGroup" data-toggle="buttons">
                    {organisationDetails}
                </div>
            </div>
        </div>;

        this.setState({
            organisations: orgDivs,
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
            errors.push("Year cannot be blank.");
        }
        if (this.state.semester === "") {
            errors.push("Semester cannot be blank.");
        }
        if (this.state.provider === "") {
            errors.push("Provider cannot be blank.");
        }
        if (this.state.directoryid === 0) {
            errors.push("Organisation cannot be blank.");
        }
        return errors;
    }

}
export {CourseForm};
