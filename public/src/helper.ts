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

export function slugify(str: string): string {

    str = str.replace(/^\s+|\s+$/g, "").toLowerCase();

    // Remove accents, swap ñ for n, etc
    const from = "ÁÄÂÀÃÅČÇĆĎÉĚËÈÊẼĔȆÍÌÎÏŇÑÓÖÒÔÕØŘŔŠŤÚŮÜÙÛÝŸŽáäâàãåčçćďéěëèêẽĕȇíìîïňñóöòôõøðřŕšťúůüùûýÿžþÞĐđßÆa·/_,:;";
    const to   = "AAAAAACCCDEEEEEEEEIIIINNOOOOOORRSTUUUUUYYZaaaaaacccdeeeeeeeeiiiinnooooooorrstuuuuuyyzbBDdBAa------";
    for (let i = 0 ; i < from.length ; i++) {
        str = str.replace(new RegExp(from.charAt(i), "g"), to.charAt(i));
    }

    // Remove invalid chars, replace whitespace by dashes, collapse dashes
    str = str.replace(/[^a-z0-9 -]/g, "").replace(/\s+/g, "-").replace(/-+/g, "-");

    return str;
}
