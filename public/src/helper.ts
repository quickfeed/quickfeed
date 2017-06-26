class ArrayHelper {

    public static find<T>(array: T[], predicate: (element: T, index: number, array: T[]) => boolean): T | null {
        for (let i = 0; i < array.length; i++) {
            const cur = array[i];
            if (predicate.call(array, cur, i, array)) {
                return cur;
            }
        }
        return null;
    }
}

export { ArrayHelper };
