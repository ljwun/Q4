export const Q4_FRONTEND_BASE_URL = process.env.Q4_FRONTEND_BASE_URL ? URL.parse(process.env.Q4_FRONTEND_BASE_URL) : null
export const BACKEND_API_BASE_URL = Q4_FRONTEND_BASE_URL ? new URL("/api", Q4_FRONTEND_BASE_URL).toString() : "/api"
