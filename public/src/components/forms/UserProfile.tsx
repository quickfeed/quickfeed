import * as React from "react";
import { UserManager } from "../../managers";
import { IUser } from "../../models";

interface IUserProfileProps {
    userMan: UserManager;
    onEditStop: () => void;
}

interface IUserProfileState {
    curUser?: IUser;
    editMode: boolean;
}

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
        let message: JSX.Element | undefined;
        if (!this.props.userMan.isValidUser(curUser)) {
            message = <div>
                It looks like your user is missing some information, please fill it out before continuing
                </div>;
        }

        const button = this.state.editMode ?
            <button
                className="btn btn-primary"
                disabled={message ? true : false}
                onClick={() => { this.stopEditing(); }}>
                Save
            </button>
            : <button className="btn btn-primary" onClick={() => { this.setState({ editMode: true }); }}>Edit</button>;
        return <div>
            {message}
            <div>
                <div className="profileElement">Firstname:</div>
                {this.renderValue(curUser, "firstname")}
            </div>
            <div>
                <div className="profileElement">Lastname:</div>
                {this.renderValue(curUser, "lastname")}
            </div>
            <div>
                <div className="profileElement">Email:</div>
                {this.renderValue(curUser, "email")}
            </div>
            <div>
                <div className="profileElement">StudentNumber:</div>
                {this.renderValue(curUser, "studentnr")}
            </div>
            {button}
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

    public renderValue(obj: any, field: string) {
        if (this.state.editMode) {
            return <input name={field} type="text" value={obj[field]} onChange={(e) => this.handleChange(e)} />;
        } else {
            return <span>{obj[field]}</span>;
        }
    }

    private copy<T extends {}>(val: T): T {
        const newEle: any = {};
        for (const a of Object.keys(val)) {
            newEle[a] = (val as any)[a];
        }
        return newEle;
    }

    private handleChange(event: React.ChangeEvent<HTMLInputElement>) {
        const name = event.target.name;
        const curUser = this.state.curUser;
        if (curUser) {
            const newUser: IUser = this.copy(curUser);
            (newUser as any)[name] = event.target.value;
            this.setState({
                curUser: newUser,
            });
        }
    }
}
