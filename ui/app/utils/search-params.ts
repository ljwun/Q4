export class EnhancedSearchParams {
    private params: { [key: string]: string | string[] | undefined };

    constructor(searchParams: { [key: string]: string | string[] | undefined }) {
        this.params = searchParams;
    }

    get(key: string): string | undefined {
        let value = this.params[key];
        if (Array.isArray(value)) {
            value = value[0];
        }
        return value;
    }

    getString(key: string, defaultValue?: string): string | undefined {
        return this.get(key) || defaultValue;
    }

    getNumber(key: string, defaultValue?: number): number | undefined {
        const value = this.get(key);
        const num = Number(value);
        return isNaN(num) ? defaultValue : num;
    }

    getBoolean(key: string, defaultValue?: boolean): boolean | undefined {
        const value = this.get(key);
        return value === 'true' ? true : value === 'false' ? false : defaultValue;
    }

    getDate(key: string, defaultValue?: Date ): Date | undefined {
        const value = this.get(key);
        return value ? new Date(value) : defaultValue;
    }
}
