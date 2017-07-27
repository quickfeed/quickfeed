import * as React from "react";
import { ILogEntry, LogLevel } from "../../managers/LogManager";

export interface IPageInfoProps {
    entry?: ILogEntry;
}

export class PageInfo extends React.Component<IPageInfoProps, {}> {
    public render() {
        console.log("PageInfoUpdate");
        console.log(this.props);
        const e = this.props.entry;
        if (!e) {
            return <div></div>;
        }
        return <div className={"topinfo alert " + this.getLevel(e)}>
            This is a message: "{e.message}"
                date: {e.date.toLocaleDateString()}
            sender: {e.sender}
            level: {e.logLevel}
        </div>;
    }

    private getLevel(entry: ILogEntry): string {
        switch (entry.logLevel) {
            default:
            case LogLevel.verbose:
            case LogLevel.info:
                return "alert-info";
            case LogLevel.warning:
                return "alert-warning";
            case LogLevel.error:
            case LogLevel.critical:
                return "alert-danger";
        }
    }
}
