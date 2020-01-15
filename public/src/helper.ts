/**
 * Binds a this variable to a function because of how Javascript works.
 * @param thisVar The this variable to be used
 * @param func The func to bind this to
 */
export function bindFunc<T>(thisVar: any, func: (props: T) => JSX.Element): (props: T) => JSX.Element {
    const temp = {
        [func.name]: (props: T) => func.call(thisVar, props),
    };
    return temp[func.name];
}

/**
 * Type for be able to use React Props in function
 */
export type RProp<T> = { children?: JSX.Element | string } & T;

/**
 * Performs a shallow copy on an object
 */
export function copy<T extends {}>(val: T): T {
    const newEle: any = {};
    for (const a of Object.keys(val)) {
        newEle[a] = (val as any)[a];
    }
    return newEle;
}

export function formatDate(str: string | Date): string {
    const dateOptions = {
        weekday: "short",
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "numeric",
        hour12: false,
    };
    const date = str instanceof Date ? str : new Date(str);
    return date.toLocaleString("no-NO", dateOptions);
}
