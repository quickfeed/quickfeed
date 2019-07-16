import * as React from "react";
import { ILogEntry, LogLevel } from "../../managers/LogManager";

interface IPageInfoProps {
    entry?: ILogEntry;
    onclose: () => void;
}

export class PageInfo extends React.Component<IPageInfoProps> {

    public render() {
        const e = this.props.entry;
        if (!e) {
            return <div></div>;
        }
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
        </div>;
    }

    // TODO(meling) why do we need both getName() and getLevel(); almost the same.
    // Can we remove one of them. And do we need to alert the user with Info level?
    // I think it is enough to have just one level: Error.

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
