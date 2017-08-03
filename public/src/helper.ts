export function bindFunc<T>(thisVar: any, func: (props: T) => JSX.Element): (props: T) => JSX.Element {
    const temp = {
        [func.name]: (props: T) => func.call(thisVar, props),
    };
    return temp[func.name];
}

export type RProp<T> = { children?: JSX.Element | string } & T;

export function copy<T extends {}>(val: T): T {
    const newEle: any = {};
    for (const a of Object.keys(val)) {
        newEle[a] = (val as any)[a];
    }
    return newEle;
}
