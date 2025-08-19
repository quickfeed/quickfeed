import { createOvermindMock } from "overmind";
import { config } from "../overmind";
import { create } from "@bufbuild/protobuf";
import { TimestampSchema } from "@bufbuild/protobuf/wkt";
export const initializeOvermind = (state, mockedApi) => {
    const overmind = createOvermindMock(config, {
        global: {
            api: mockedApi
        }
    }, initialState => {
        Object.assign(initialState, state);
    });
    Object.assign(overmind.effects.global.api, mockedApi);
    return overmind;
};
export function mock(_method, mockFn) {
    return async function (...args) {
        return mockFn(...args);
    };
}
const toTimestamp = (date) => {
    const seconds = BigInt(Math.floor(date.getTime() / 1000));
    const nanos = (date.getTime() % 1000) * 1e6;
    return create(TimestampSchema, { seconds, nanos });
};
const dateSet = () => {
    const date = new Date();
    return {
        date,
        year: date.getFullYear(),
        month: date.getMonth(),
        dayOfTheMonth: date.getDate(),
        dayOfTheWeek: date.getDay(),
        hours: date.getHours(),
        minutes: date.getMinutes(),
        seconds: date.getSeconds(),
        milliseconds: date.getMilliseconds(),
    };
};
export const timeStamp = ({ years, months, days, hours } = {}) => {
    const set = dateSet();
    const add = (value, mentor) => {
        return value + (mentor ?? 0);
    };
    const year = add(set.year, years);
    const month = add(set.month, months);
    const day = add(set.dayOfTheMonth, days);
    const hour = add(set.hours, hours);
    return toTimestamp(new Date(year, month, day, hour));
};
