'use client';

import { useEffect, useState } from 'react'
import { BACKEND_API_BASE_URL } from '@/app/constants'
import createClient from "openapi-fetch";
import type { paths } from "@/app/openapi/openapi";
import Link from 'next/link';
import { makeCallbackUrl } from '@/app/constants';
import { useRouter } from 'next/navigation';

const client = createClient<paths>({ baseUrl: BACKEND_API_BASE_URL });

export default function Callback() {
    const router = useRouter();
    const [status, setStatus] = useState<string>("登入處理中...");
    const [prompt, setPrompt] = useState<string>("");
    const [redirectLink, setRedirectLink] = useState<string | null>(null);

    useEffect(() => {
        (async () => {
            const searchParam = new URLSearchParams(window.location.search);
            const redirectUrl = searchParam.get('redirect_url');
            async function handleAuthCode() {
                const code = searchParam.get('code');
                const state = searchParam.get('state');
                if (!code || !state || !redirectUrl) {
                    setStatus("登入失敗，缺少部分必要參數!");
                    return;
                }
                const { error, response } = await client.GET("/auth/callback", {
                    params: {
                        query: {
                            code: code,
                            state: state,
                            redirect_url: makeCallbackUrl(new URL(redirectUrl)),
                        },
                    },
                });
                if (error) {
                    setStatus("登入失敗，授權碼交換失敗!");
                    console.error('Error during login:', error);
                    return;
                }
                setStatus("登入成功，正在導向原始網頁...");
                const location = response?.headers.get('Location')
                if (location) {
                    router.replace(location);
                }else {
                    router.replace("/");
                }
                router.refresh();
            };
            await handleAuthCode();
            setPrompt("十五秒後將會自動跳轉:");
            setRedirectLink(redirectUrl)
            // setTimeout(() => {
            //     if (!redirectUrl) {
            //         setPrompt("跳轉失敗，缺少部分必要參數");
            //         return
            //     }
            //     window.location.href = redirectUrl;
            // }  , 15000);
        })();
    }, [router]);

    return (
        <>
            <div>{status}</div>
            <div>
                {prompt}
                {redirectLink ? <Link href={redirectLink}>{redirectLink}</Link> : null}
            </div>
        </>
    );
}
