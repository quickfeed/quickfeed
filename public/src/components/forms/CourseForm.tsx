import * as React from "react";
import { BootstrapButton } from "../../components";
import { CourseManager } from "../../managers/CourseManager";
import { NavigationManager } from "../../managers/NavigationManager";
import { Course, Status, User, Void } from "../../../proto/ag/ag_pb";

interface ICourseFormProps {
    className?: string;
    courseMan: CourseManager;
    navMan: NavigationManager;
    pagePath: string;
    courseData?: Course; // for editing an existing course
    curUser?: User;  // to make current user info accessible
    providers: string[];
}

interface ICourseFormState {
    name: string;
    code: string;
    tag: string;
    year: string;
    slipdays: string;
    provider: string;
    orgid: number;
    orgname: string;
    errorFlash: JSX.Element | null;
    userMessage: JSX.Element | null;
    success: number;
    clicked: boolean;
}

export class CourseForm<T> extends React.Component<ICourseFormProps, ICourseFormState> {

    constructor(props: any) {
        super(props);
        this.state = {
            name: this.props.courseData?.getName() ?? "",
            code: this.props.courseData?.getCode() ?? "",
            tag: this.props.courseData?.getTag() ?? "",
            year: this.props.courseData?.getYear().toString() ?? "",
            slipdays: this.props.courseData?.getSlipdays().toString() ?? "",
            orgname: "",
            provider: "github",
            orgid: this.props.courseData ? this.props.courseData.getOrganizationid() : 0,
            errorFlash: null,
            userMessage: null,
            success: 0,
            clicked: false,
        };
    }

    public render() {
        const getTitleText: string = this.props.courseData ? "Edit Course" : "Create New Course";
        return (
            <div className="container">
                <div className="row">
                    <div className="form-header col-sm-12"><h1>{getTitleText}</h1></div>
                </div>
                <div className="row">{this.state.errorFlash}</div>
                <form className={this.props.className ? this.props.className : ""}
                    onSubmit={(e) => this.handleFormSubmit(e)}>
                    <div className="form-group" id="organisation-container">
                        <div className="col-sm-10">
                            {this.renderInfo()}
                        </div>
                    </div>
                    <div className="row spacefix">
                        {this.props.courseData ? null : this.courseByName()}
                    </div>
                    <div className="row spacefix">
                        {this.renderFormController("Name:",
                            "Enter course name",
                            "name",
                            this.state.name,
                            (e) => this.handleInputChange(e))}
                        {this.renderFormController("Code:",
                            "Enter course code",
                            "code",
                            this.state.code,
                            (e) => this.handleInputChange(e))}
                    </div>
                    <div className="row spacefix">
                        {this.renderFormController("Year:",
                            "Enter year",
                            "year",
                            this.state.year,
                            (e) => this.handleInputChange(e))}
                        {this.renderFormController("Tag:",
                            "Enter semester",
                            "tag",
                            this.state.tag,
                            (e) => this.handleInputChange(e))}
                    </div>
                    <div className="row spacefix">
                        {this.renderFormController("Slip days:",
                            "Slip days",
                            "slipdays",
                            this.state.slipdays,
                            (e) => this.handleInputChange(e))}
                    </div>
                    <div className="row spacefix">
                        <div className="form-button col-sm-12">
                            <BootstrapButton classType="primary" type="submit">
                                {this.setButtonString()}
                            </BootstrapButton>
                        </div>
                    </div>
                </form>
            </div>
        );
    }

    private renderInfo(): JSX.Element {
        const gitMsg: JSX.Element =
            <div>
                <p>For each new semester of a course, Autograder requires a new GitHub organization.
            This is to keep the student roster for the different runs of the course separate.</p>

                <p><a href="https://github.com/account/organizations/new"
                    target="_blank">Create an organization for your course</a>.
                  The course organization must allow private repositories. </p>

                <p>Autograder will create a following repositories for you:</p>

                <div>
                    <ul>
                        <li>info</li>
                        <li>assignments</li>
                        <li>tests</li>
                    </ul>
                </div>

                <p>Please read <a
                    href="https://github.com/autograde/quickfeed/blob/master/doc/teacher.md"
                    target="_blank">the documentation</a> for further
                   instructions on how to work with the various repositories.</p>
            </div>;
        return gitMsg;
    }

    private renderFormController(
        title: string,
        placeholder: string,
        name: string,
        value: any,
        onChange: (e: React.ChangeEvent<HTMLInputElement>) => void,
    ) {
        return <div className="col-sm-6"><div className="input-group">
            <label className="input-group-addon addon-mini" htmlFor="name">{title}</label>
            <input type="text" className="form-control"
                id={name}
                placeholder={placeholder}
                name={name}
                value={value}
                onChange={onChange}
            />
        </div></div>;
    }

    private async handleFormSubmit(e: React.FormEvent<any>) {
        e.preventDefault();
        const errors: string[] = this.courseValidate();
        this.setState({
            clicked: true,
        });
        if (errors.length > 0) {
            const flashErrors = this.getFlashErrors(errors);
            this.setState({
                errorFlash: flashErrors,
            });
        } else {
            const result = this.props.courseData ?
                await this.updateCourse(this.props.courseData.getId()) : await this.createNewCourse();

            if ((result instanceof Status) && (result.getCode() > 0)) {
                const errMsg = result.getError();
                const serverErrors: string[] = [];
                serverErrors.push(errMsg);
                const flashErrors = this.getFlashErrors(serverErrors);
                this.setState({
                    errorFlash: flashErrors,
                });
            } else {
                const redirectTo: string = this.props.pagePath + "/courses";
                this.props.navMan.navigateTo(redirectTo);
            }
        }
        this.setState({
            clicked: false,
        });
    }

    private async updateCourse(courseId: number): Promise<Void | Status> {
        const newCourse = new Course();
        newCourse.setId(courseId);
        newCourse.setName(this.state.name);
        newCourse.setCode(this.state.code);
        newCourse.setTag(this.state.tag);
        newCourse.setYear(parseInt(this.state.year, 10));
        newCourse.setProvider(this.state.provider);
        newCourse.setOrganizationid(this.state.orgid);
        newCourse.setSlipdays(parseInt(this.state.slipdays, 10));

        return this.props.courseMan.updateCourse(newCourse);
    }

    private async createNewCourse(): Promise<Course | Status> {
        const courseData = new Course();
        courseData.setName(this.state.name);
        courseData.setCode(this.state.code);
        courseData.setTag(this.state.tag);
        courseData.setYear(parseInt(this.state.year, 10));
        courseData.setProvider(this.state.provider);
        courseData.setOrganizationid(this.state.orgid);
        courseData.setSlipdays(parseInt(this.state.slipdays, 10));

        return this.props.courseMan.createNewCourse(courseData);
    }

    private handleInputChange(e: React.FormEvent<any>) {
        const target: any = e.target;
        const value = target.type === "checkbox" ? target.checked : target.value;
        const name = target.name as "name";

        this.setState({
            [name]: value,
        });
    }

    private async getOrgByName(orgName: string) {
        const accessLinkString = "https://github.com/organizations/" + orgName + "/settings/oauth_application_policy";
        const accessLink = <a href={accessLinkString} target="_blank">third-party access</a>;
        const result = await this.props.courseMan.getOrganization(orgName);
        if (result instanceof Status) {
            this.setState({
                success: 2,
            });
            // if error message has code 9, it is supposed to be shown to user
            if (result.getCode() === 9) {
                this.setState({ userMessage: <span>{result.getError()}</span> });
            } else {
                this.setState({
                    userMessage: <span>course organization must remove restrictions for {accessLink}</span>
                })
            }

        } else {
            this.setState({
                userMessage: <span>Organization found</span>,
                success: 1,
                orgid: result.getId(),
            });
        }
    }

    private courseByName() {
        return <div className="col-sm-12">
            <div className="input-group orgform">
                <label className="input-group-addon">Organization:</label>
                <input type="text"
                    className="form-control"
                    id="orgname"
                    placeholder="Course organization name"
                    name="orgname"
                    value={this.state.orgname}
                    onChange={(e) => this.handleInputChange(e)}
                /> <span className="input-group-btn"><button className="btn btn-primary" type="button"
                    onClick={(e) => this.getOrgByName(this.state.orgname)}
                >Find</button></span></div>
            <label className="control-label col-sm-1" htmlFor="name"></label>
            <div id="message" className="col-sm-11" >
                <span className={this.setMessageIcon()}>
                </span>  {this.state.userMessage}</div>
        </div>;
    }

    private courseValidate(): string[] {
        const errors: string[] = [];
        if (this.state.name === "") {
            errors.push("Course name cannot be blank");
        }
        if (this.state.code === "") {
            errors.push("Course tag cannot be blank.");
        }
        if (this.state.tag === "") {
            errors.push("Semester cannot be blank.");
        }
        if (this.state.provider === "") {
            errors.push("Unknown provider.");
        }
        const slipdays = parseInt(this.state.slipdays, 10);
        if (this.state.slipdays === "") {
            errors.push("Slip days not set");
        } else if (Number.isNaN(slipdays)) {
            errors.push("Slip days must be a number");
        }

        if (this.state.orgid === 0) {
            errors.push("Select organization for your course.");
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

    private getFlashErrors(errors: string[]): JSX.Element {
        const errorArr: JSX.Element[] = [];
        for (let i: number = 0; i < errors.length; i++) {
            errorArr.push(<li key={i}>{errors[i]}</li>);
        }
        const flash: JSX.Element =
            <div className="alert alert-danger">
                <h4> Cannot create a new course: </h4>
                <ul>
                    {errorArr}
                </ul>
            </div>;
        return flash;
    }

    private setMessageIcon(): string {
        switch (this.state.success) {
            case 1: {
                return "glyphicon glyphicon-ok";
            }
            case 2: {
                return "glyphicon glyphicon-remove";
            }
            default: {
                return "";
            }
        }
    }

    private setButtonString(): string {
        if (this.state.clicked) {
            return this.props.courseData ? "Updating" : "Creating";
        }
        return this.props.courseData ? "Update" : "Create";
    }
}
