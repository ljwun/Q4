import { NextRequest, NextResponse } from 'next/server';

export async function middleware(request: NextRequest) {
    const backendUrl = process.env.Q4_BACKEND_BASE_URL
    if (!backendUrl) {
        return NextResponse.next();
    }

    const url = new URL(request.nextUrl.pathname.slice(4), backendUrl);
    const searchParams = new URLSearchParams(request.nextUrl.search);
    url.search = searchParams.toString();
    console.log(`Proxying request to ${url.toString()}`);

    // 檢查是否為 SSE 請求
    const isSSE = request.headers.get('accept') === 'text/event-stream'

    // 如果不是 SSE 請求，則直接轉發
    if (!isSSE) {
        return NextResponse.rewrite(url);
    }

    // 如果是 SSE 請求，則使用 fetch 進行轉發
    // 複製所有請求 headers
    const requestHeaders = new Headers(request.headers)
    
    try {
        const response = await fetch(url, {
            method: request.method,
            headers: requestHeaders,
            body: request.body,
        })

        // 複製所有響應 headers
        const responseHeaders = new Headers()
        response.headers.forEach((value, key) => {
            // 某些 headers 可能會被 Next.js 覆蓋或忽略
            try {
                responseHeaders.set(key, value)
            } catch (e) {
                console.warn(`Cannot set header: ${key}`, e)
            }
        })

        // 創建新的響應
        return new NextResponse(response.body, {
            status: response.status,
            statusText: response.statusText,
            headers: responseHeaders,
        })
    } catch (error) {
        console.error('Proxy error:', error)
        return NextResponse.error()
    }
}

export const config = {
    matcher: '/api/:path*',
};