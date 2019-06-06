import * as React from "react";
import { User } from "../../../proto/ag_pb";
import { UserManager } from "../../managers";
import { BootstrapButton } from "../../components";

interface IUserProfileProps {
    userMan: UserManager;
    onEditStop: () => void;
}

interface IUserProfileState {
    curUser?: User;
    editMode: boolean;
}

type RWrap<T> = (props: T) => JSX.Element;

export class UserProfile extends React.Component<IUserProfileProps, IUserProfileState> {
    constructor(props: IUserProfileProps, context: any) {
        super(props, context);
        const curUser = props.userMan.getCurrentUser();
        if (curUser) {
            this.state = {
                curUser,
                editMode: !props.userMan.isValidUser(curUser),
            };
        }
    }

    public render() {
        if (!this.state.curUser) {
            return <h1>User not logged in</h1>;
        }
        const curUser = this.state.curUser;
        return <div>
            <div className="row container center-block">
                <div className="col-md-3">
                    {this.renderUserInfoBox(curUser)}
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
                To sign up, please complete the form below.
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

        if (this.state.curUser && this.props.userMan.isValidUser(this.state.curUser)) {
            await this.props.userMan.updateUser(this.state.curUser);
            await this.props.userMan.checkUserLoggedIn();
            const curUser = this.props.userMan.getCurrentUser();
            if (curUser) {
                this.setState({ editMode: false, curUser });
            } else {
                this.setState({ editMode: false });
            }
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
                console.log("UserProfile: unknown field name when rendering user fields: " + field);
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
        const curUser = this.state.curUser;
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
                    console.log("UserProfile change event gets some other value: " + name);
                    break;
                }
            }

            this.setState({
                curUser: newUser,
            });
        }
    }
}
