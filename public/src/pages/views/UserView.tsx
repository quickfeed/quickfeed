import * as React from "react";
import { DynamicTable } from "../../components";
import { IUser } from "../../models";

interface IUserViewerProps{
    users: IUser[];
} 

class UserView extends React.Component<IUserViewerProps, undefined> {
    render(){
        return <DynamicTable 
            header={["ID","First name", "Last name", "Email", "StudentID"]} 
            data={this.props.users} 
            selector={(item: IUser) => [item.id.toString(), item.firstName, item.lastName, item.email, item.personId.toString()]} 
            >
        </DynamicTable>
    }
}

export {UserView}