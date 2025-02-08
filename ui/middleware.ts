import { NextRequest, NextResponse } from 'next/server';

export function middleware(request: NextRequest) {
    const backendUrl = process.env.Q4_BACKEND_BASE_URL
    if (backendUrl) {
        const url = new URL(request.nextUrl.pathname.slice(4), backendUrl);
        const searchParams = new URLSearchParams(request.nextUrl.search);
        url.search = searchParams.toString();
        console.log(`Proxying request to ${url.toString()}`);
        return NextResponse.rewrite(url);
    }
    return NextResponse.next();
}

export const config = {
    matcher: '/api/:path*',
};