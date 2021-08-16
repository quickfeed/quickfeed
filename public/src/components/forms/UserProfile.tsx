import * as React from "react";
import { User } from "../../../proto/ag/ag_pb";
import { BootstrapButton } from "../../components";
import { UserManager } from "../../managers";

interface IUserProfileProps {
    userMan: UserManager;
    onEditStop: () => void;
}

interface IUserProfileState {
    editMode: boolean;
    toggle: boolean;
}

export class UserProfile extends React.Component<IUserProfileProps, IUserProfileState> {
    private curUser: User | null;
    constructor(props: IUserProfileProps, context: any) {
        super(props, context);
        this.curUser = props.userMan.getCurrentUser();
        if (this.curUser) {
            this.state = {
                editMode: !props.userMan.isValidUser(this.curUser),
                toggle: false,
            };
        }
    }

    public render() {
        if (!this.curUser) {
            return <h1>User not logged in</h1>;
        }
        return <div>
            <div className="row container center-block">
                <div className="col-md-3">
                    {this.renderUserInfoBox(this.curUser)}
                </div>
                <div className="col-md-9">
                    <h3>There is currently nothing important to note</h3>
                </div>
            </div>
        </div >;
    }

    public renderUserInfoBox(curUser: User): JSX.Element {
        let message: JSX.Element | undefined;
        if (!this.props.userMan.isValidUser(curUser)) {
            message = <div style={{ color: "red" }}>
                <p>To sign up, please complete the form below.</p>
                <p>Use your <i>real name</i> as it appears on Canvas.</p>
                <p>If your name does not match any names on Canvas, you will not be granted access.</p>
                </div>;
        }

        return <div className="well">
            <h3>Your information</h3>
            {message}
            {this.renderField("name", curUser, "Name")}
            {this.renderField("email", curUser, "Email")}
            {this.renderField("studentid", curUser, "Student id")}
            {this.renderSaveButton(message !== undefined, this.state.editMode)}
        </div>;
    }

    public renderSaveButton(disabled: boolean, editMode: boolean) {

        if (editMode) {
            return <BootstrapButton
                classType="primary"
                disabled={disabled}
                onClick={() => { this.stopEditing(); }}>
                Save
            </BootstrapButton>;
        } else {
            return <BootstrapButton
                classType="primary"
                onClick={() => { this.setState({ editMode: true }); }}>
                Edit
            </BootstrapButton>;
        }
    }

    public renderField(value: string, obj: any, children?: JSX.Element | string): JSX.Element {
        return <div className="field-box">
            <b>{children}</b>
            {this.renderValue(value, obj)}
        </div>;
    }

    public async stopEditing() {
        if (this.curUser && this.props.userMan.isValidUser(this.curUser)) {
            await this.props.userMan.updateUser(this.curUser);
            await this.props.userMan.checkUserLoggedIn();
            const curUser = this.props.userMan.getCurrentUser();
            if (curUser) {
                this.curUser = curUser;
            }
            this.setState({ editMode: false});
            this.props.onEditStop();
        }
    }

    public renderValue(field: string, obj: any) {
        // grpc class has no public fields, to use a right getter check what value is rendering
        let renderString = "";
        switch (field) {
            case "name": {
                renderString = (obj as User).getName();
                break;
            }
            case "email": {
                renderString = (obj as User).getEmail();
                break;
            }
            case "studentid": {
                renderString = (obj as User).getStudentid();
                break;
            }
            default: {
                break;
            }
        }

        if (this.state.editMode) {
            return <input
                className="form-control"
                name={field}
                type="text"
                value={renderString}
                onChange={(e) => this.handleChange(e)} />;
        } else {
            return <span>{renderString}</span>;
        }
    }

    private handleChange(event: React.ChangeEvent<HTMLInputElement>) {
        const name = event.target.name;
        const curUser = this.curUser;
        if (curUser) {
            const newUser: User = new User();
            newUser.setId(curUser.getId());
            newUser.setName(curUser.getName());
            newUser.setStudentid(curUser.getStudentid());
            newUser.setEmail(curUser.getEmail());
            newUser.setAvatarurl(curUser.getAvatarurl());
            newUser.setIsadmin(curUser.getIsadmin());
            newUser.setRemoteidentitiesList(curUser.getRemoteidentitiesList());
            newUser.setEnrollmentsList(curUser.getEnrollmentsList());
            switch (name) {
                case "name": {
                    newUser.setName(event.target.value);
                    break;
                }
                case "email": {
                    newUser.setEmail(event.target.value);
                    break;
                }
                case "studentid": {
                    newUser.setStudentid(event.target.value);
                    break;
                }
                default: {
                    break;
                }
            }

            this.curUser = newUser;
            this.setState({toggle: !this.state.editMode});
        }
    }
}
