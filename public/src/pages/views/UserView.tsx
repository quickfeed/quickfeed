import * as React from "react";
import {DynamicTable, Search} from "../../components";
import {IUser} from "../../models";

interface IUserViewerProps {
    users: IUser[];
    addSearchOption?: boolean;
}

interface IUserViewerState {
    users: IUser[];
}

class UserView extends React.Component<IUserViewerProps, IUserViewerState> {
    constructor(props: any) {
        super(props);
        this.state = {
            users: this.props.users,
        };
    }

    public render() {
        let searchForm: JSX.Element | null = null;
        if (this.props.addSearchOption) {
            const searchIcon: JSX.Element = <span className="input-group-addon">
            <i className="glyphicon glyphicon-search"></i>
        </span>;
            searchForm = <Search className="input-group"
                                 addonBefore={searchIcon}
                                 placeholder="Search for students"
                                 onChange={(query) => this.handleOnchange(query)}
            />;
        }
        return (
            <div>
                {searchForm}
                <DynamicTable
                    header={["ID", "First name", "Last name", "Email", "StudentID"]}
                    data={this.state.users}
                    selector={(item: IUser) => [
                        item.id.toString(),
                        item.firstName,
                        item.lastName,
                        item.email,
                        item.personId.toString(),
                    ]}
                />
            </div>);
    }

    private handleOnchange(query: string): void {
        query = query.toLowerCase();
        const filteredData: IUser[] = [];
        this.props.users.forEach((user) => {
            if (user.firstName.toLowerCase().indexOf(query) !== -1
                || user.lastName.toLowerCase().indexOf(query) !== -1
                || user.email.toLowerCase().indexOf(query) !== -1
                || user.personId.toString().indexOf(query) !== -1
            ) {
                filteredData.push(user);
            }
        });

        this.setState({
            users: filteredData,
        });
    }
}

export {UserView};
