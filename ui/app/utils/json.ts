export function reParseJSON<Type>(obj: NonNullable<Type>, reviver?: (this: unknown, key: string, value: unknown) => unknown): Type {
    return JSON.parse(JSON.stringify(obj), reviver) as Type;
}

export function dateReviver(key: string, value: unknown): unknown {
    if (typeof value === "string") {
        // 使用 ISO8601 日期格式來檢查
        // 參考: https://stackoverflow.com/questions/12756159/regex-and-iso8601-formatted-datetime
        const dateRegex = /^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(\.\d+)?(([+-]\d\d:\d\d)|Z)?$/i;
        if (dateRegex.test(value)) {
            return new Date(value);
        }
    }
    return value;
}