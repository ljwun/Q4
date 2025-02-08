export function serializeDeepObject(obj: unknown, prefix: string = ''): string {
    if (!obj || typeof obj !== 'object') {
        return '';
    }

    return Object.keys(obj)
        .map(key => {
            const value = (obj as { [key: string]: unknown })[key];
            const keyPrefix = prefix ? `${prefix}[${key}]` : key;

            if (value === null || value === undefined) {
                return '';
            }

            if (Array.isArray(value)) {
                return value
                    .map((item, index) => {
                        if (typeof item === 'object') {
                            return serializeDeepObject(item, `${keyPrefix}[${index}]`);
                        }
                        return `${keyPrefix}[${index}]=${encodeURIComponent(item)}`;
                    })
                    .filter(Boolean)
                    .join('&');
            }

            if (value instanceof Date) {
                return `${keyPrefix}=${encodeURIComponent(value.toISOString())}`;
            }

            if (typeof value === 'object') {
                return serializeDeepObject(value, keyPrefix);
            }

            if (typeof value === 'string' || typeof value === 'number' || typeof value === 'boolean') {
                return `${keyPrefix}=${encodeURIComponent(value)}`;
            }
            console.warn(`Cannot serialize unsupported type: ${typeof value}`);
            return '';
        })
        .filter(Boolean)
        .join('&');
}

// Usage example:
// const params = { 
//     filter: { 
//         name: 'John',
//         birthDate: new Date('2023-01-01'),
//      address: { city: 'NY', zip: '10001' } 
//     } 
// };
// serializeDeepObject(params)
// Result: "filter[name]=John&filter[birthDate]=2023-01-01T00%3A00%3A00.000Z&filter[address][city]=NY&filter[address][zip]=10001"
