import * as React from "react";
import { ILogEntry, LogLevel } from "../../managers/LogManager";

export interface IPageInfoProps {
    entry?: ILogEntry;
    onclose: () => void;
}

export class PageInfo extends React.Component<IPageInfoProps, {}> {
    public render() {
        console.log("PageInfoUpdate");
        console.log(this.props);
        const e = this.props.entry;
        if (!e) {
            return <div></div>;
        }
        // [{e.date.toLocaleDateString()}]
        return <div className={"topinfo alert " + this.getLevel(e)}>
            <button
                type="button"
                className="close"
                onClick={() => this.props.onclose()}>
                <span aria-hidden="true">
                    &times;
                </span>
            </button>
            <strong>{this.getName(e)}</strong>: {e.message}
        </div >;
    }

    private getName(entry: ILogEntry): string {
        switch (entry.logLevel) {
            default:
            case LogLevel.verbose:
            case LogLevel.info:
                return "Info";
            case LogLevel.warning:
                return "Warning!";
            case LogLevel.error:
            case LogLevel.critical:
                return "Error!";
        }
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
