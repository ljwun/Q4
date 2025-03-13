'use client'

import createClient from "openapi-fetch";
import { BACKEND_API_BASE_URL } from '@/app/constants'
import { components, paths, Defined } from "@/app/openapi";

type SSOProviderType = Defined<components["schemas"]["SSOProvider"]>
export type LoginMessage = {
    status: LoginStatus
    error?: string
}
export enum LoginStatus {
    loginSuccess = "loginSuccess",
    loginFailed = "loginFailed",
}

// 以popup方式開啟登入視窗執行登入
export async function login(provider: SSOProviderType, toast: (props: { title: string, description: string, variant?: "destructive" }) => void, refresher?: ()=>void) {
    const client = createClient<paths>({ baseUrl: BACKEND_API_BASE_URL });
    // 取得登入網址
    const { error, response } = await client.GET("/auth/sso/{provider}/login", {
        params: {
            path: { provider },
            query: {
                redirectUrl: new URL(`/sso/${provider}/callback`, location.origin).toString(),
            },
        },
    });
    if (response.status == 404) {
        toast({
            title: "登入失敗",
            description: "目前不支援此登入方式。",
            variant: "destructive",
        });
        return
    }else if (response.status != 200) {
        toast({
            title: "登入失敗",
            description: "無法跳轉到登入網頁，請稍後再試。",
            variant: "destructive",
        });
        console.log('Error during login:', error);
        return
    }
    const authUrl = response?.headers.get('Location')
    if (!authUrl) {
        toast({
            title: "登入失敗",
            description: "無法取得登入網址，請稍後再試。",
            variant: "destructive",
        });
        return
    } 
    // 開啟登入視窗並監聽
    const authWindow = window.open(authUrl, "authWindow");
    window.addEventListener("message", (event) => {
        if (event.origin !== location.origin) {
            return
        }
        // 檢查資料是否是LoginMessage類型
        if (!("status" in event.data)) {
            return
        }
        const message = event.data as LoginMessage
        switch (message.status) {
            case LoginStatus.loginSuccess:
                toast({
                    title: "登入成功",
                    description: "登入成功",
                });
                authWindow?.close()
                
                break
            case LoginStatus.loginFailed:
                toast({
                    title: "登入失敗",
                    description: "登入失敗，請稍後再試。",
                    variant: "destructive",
                });
                console.log('Error during login:', message.error);
                break
        }
        if (refresher) {
            refresher()
        }
    })
}