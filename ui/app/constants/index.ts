export const Q4_FRONTEND_BASE_URL = process.env.Q4_FRONTEND_BASE_URL ? URL.parse(process.env.Q4_FRONTEND_BASE_URL) : null
export const BACKEND_API_BASE_URL = Q4_FRONTEND_BASE_URL ? new URL("/api", Q4_FRONTEND_BASE_URL).toString() : "/api"

export function makeCallbackUrl(redirectUrl: URL): string {
    const url = new URL("/callback", redirectUrl.origin);
    url.searchParams.set("redirect_url", redirectUrl.toString());
    return url.toString();
}