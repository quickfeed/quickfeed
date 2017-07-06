export interface IEventData {
    target: any;
}

export interface INewEvent<T extends IEventData> {
    (event: T): void;
    info: string;
    addEventListener(listener: (event: T) => void): void;
    removeEventListener(listener: (event: T) => void): void;
}

/**
 * Creates a new event with a given name.
 * The T argument in typescript should be an interface or class implementing
 * the IEventData interface.
 * @param info The name of the event that should fire.
 * Usualy at the format {ClassName/sender}.{eventName}
 */
export function newEvent<T extends IEventData>(info: string): INewEvent<T> {
    const callbacks: Array<((event: T) => void)> = [];

    const handler = function EventHandler(event: T) {
        callbacks.map(((v) => v(event)));
    } as INewEvent<T>;

    handler.info = info;
    handler.addEventListener = (callback) => {
        callbacks.push(callback);
    };
    handler.removeEventListener = (callback) => {
        const index = callbacks.indexOf(callback);
        if (index < 0) {
            console.log(callback);
            throw Error("Event does noe exist");
        }
        callbacks.splice(index, 1);
    };
    return handler;
}
