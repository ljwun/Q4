'use client';

import { useEffect, useState } from 'react'
import { BACKEND_API_BASE_URL } from '@/app/constants'
import createClient from "openapi-fetch";
import type { paths } from "@/app/openapi/openapi";
import Link from 'next/link';
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
            async function handleAuthCode() {
                const code = searchParam.get('code');
                const state = searchParam.get('state');
                if (!code || !state ) {
                    setStatus("登入失敗，缺少部分必要參數!");
                    return;
                }
                const { data, error, response } = await client.GET("/auth/callback", {
                    params: {
                        query: {
                            code: code,
                            state: state,
                        },
                    },
                });
                if (response.status == 200) {
                    setStatus("登入成功，正在導向原始網頁...");
                    const location = data?.urlBeforeLogin
                    if (location) {
                        router.replace(location);
                    }else {
                        router.replace("/");
                    }
                    router.refresh();
                }else {
                    setStatus("登入失敗，授權碼交換失敗!");
                    console.error('Error during login:', error);
                }
            };
            await handleAuthCode();
            setPrompt("十五秒後將會自動跳轉:");
            setRedirectLink("/")
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
