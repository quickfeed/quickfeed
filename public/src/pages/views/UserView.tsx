import * as React from "react";
import { DynamicTable } from "../../components";
import { IUser } from "../../models";

class UserViewer extends React.Component<any, undefined> {
    render(){
        return <DynamicTable 
            header={["ID","First name", "Last name", "Email", "StudentID"]} 
            data={this.props.users} 
            selector={(item: IUser) => [item.id.toString(), item.firstName, item.lastName, item.email, item.personId.toString()]} 
            >
        </DynamicTable>
    }
}

export {UserViewer}