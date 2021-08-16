import { IEventData, newEvent } from "../event";

export interface ILogEntry {
    date: Date;
    message: string;
    logLevel: LogLevel;
    sender: string;
}

export enum LogLevel {
    verbose,
    info,
    warning,
    error,
    critical,
}

interface ILogEvent extends IEventData {
    entry: ILogEntry;
}

export interface ILogger {
    log(message: string, showUser?: boolean): void;
    warn(message: string, showUser?: boolean): void;
    error(message: string, showUser?: boolean): void;
}

export interface ILoggerServer {
    createLogger(name: string): ILogger;
    pushEntry(message: string, logLevel: LogLevel, sender: string, showUser: boolean): void;
}

export class LogClient implements ILogger {
    private logger: ILoggerServer;
    private name: string;
    constructor(name: string, logger: ILoggerServer) {
        this.name = name;
        this.logger = logger;
    }

    public log(message: string, showUser: boolean = false): void {
        this.logger.pushEntry(message, LogLevel.info, this.name, showUser);
    }

    public warn(message: string, showUser: boolean = false) {
        this.logger.pushEntry(message, LogLevel.warning, this.name, showUser);
    }

    public error(message: string, showUser: boolean = false) {
        this.logger.pushEntry(message, LogLevel.error, this.name, showUser);
    }
}

// tslint:disable-next-line:max-classes-per-file
export class LogManager implements ILogger, ILoggerServer {
    public onshowuser = newEvent<ILogEvent>("LogManager.onshowuser");
    public name: string = "LogManager";

    private allLog: ILogEntry[] = [];

    public log(message: string, showUser: boolean = false): void {
        this.pushEntry(message, LogLevel.info, this.name, showUser);
    }

    public warn(message: string, showUser: boolean = false) {
        this.pushEntry(message, LogLevel.warning, this.name, showUser);
    }

    public error(message: string, showUser: boolean = false) {
        this.pushEntry(message, LogLevel.error, this.name, showUser);
    }

    public createLogger(name: string): ILogger {
        return new LogClient(name, this);
    }

    public pushEntry(message: string, logLevel: LogLevel, sender: string, showUser: boolean = false) {
        const entry: ILogEntry = { date: new Date(), message, logLevel, sender };
        this.allLog.push(entry);
        if (showUser) {
            this.onshowuser({ target: this, entry });
        }
    }
}
