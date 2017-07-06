export class ArrayHelper {

    public static join<T, T2>(
        array1: T[],
        array2: T2[],
        callback: (ele1: T, ele2: T2) => boolean): Array<{ ele1: T, ele2: T2 }> {
        const returnObj: Array<{ ele1: T, ele2: T2 }> = [];
        for (const ele1 of array1) {
            for (const ele2 of array2) {
                if (callback(ele1, ele2)) {
                    returnObj.push({ ele1, ele2 });
                }
            }
        }
        return returnObj;
    }

    public static find<T>(array: T[], predicate: (element: T, index: number, array: T[]) => boolean): T | null {
        for (let i = 0; i < array.length; i++) {
            const cur = array[i];
            if (predicate.call(array, cur, i, array)) {
                return cur;
            }
        }
        return null;
    }

    public static async mapAsync<inT, outT>(
        array: inT[],
        callback: (
            element: inT,
            index: number,
            array: inT[]) => Promise<outT>): Promise<outT[]> {
        const newArray: outT[] = [];
        for (let i = 0; i < array.length; i++) {
            newArray.push(await callback(array[i], i, array));
        }
        return newArray;
    }
}
