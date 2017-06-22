interface IEventData{
    target: any;
}

interface INewEvent<T extends IEventData>{
    (event: T): void;
    info: string;
    addEventListener(listener: (event: T) => void): void;
    removeEventListener(listener: (event: T) => void): void;
}

function newEvent<T extends IEventData>(info: string): INewEvent<T> {
    let callbacks: ((event: T) => void)[] = [];
    let handler = function EventHandler(event: T){
        callbacks.map(v => v(event));
    } as INewEvent<T>;
    handler.info = info;
    handler.addEventListener = (callback) => {
        callbacks.push(callback);
    }
    handler.removeEventListener = (callback) => {
        let index = callbacks.indexOf(callback);
        if (index < 0){
            console.log(callback);
            throw Error("Event does noe exist");
        }
        callbacks.splice(index, 1);
    }
    return handler;
}

export {IEventData, INewEvent, newEvent}