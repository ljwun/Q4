'use client';

import { useEffect } from 'react'
import { BACKEND_API_BASE_URL } from '@/app/constants'
import createClient from "openapi-fetch";
import type { components, paths, Defined } from "@/app/openapi";
import { useRouter } from 'next/navigation';
import {LoginMessage, LoginStatus} from '@/app/components/user/login';

type SSOProviderType = Defined<components["schemas"]["SSOProvider"]>
const client = createClient<paths>({ baseUrl: BACKEND_API_BASE_URL });

export default function Link(props: { params: Promise<{ provider: SSOProviderType }> }) {
    const router = useRouter();

    useEffect(() => {
        (async () => {
            const params = await props.params;
            const searchParam = new URLSearchParams(window.location.search);
            async function handleAuthCode() {
                const code = searchParam.get('code');
                const state = searchParam.get('state');
                if (!code || !state ) {
                    window.opener.postMessage({status: LoginStatus.loginFailed, error: "缺少必要參數"} as LoginMessage, window.location.origin);
                    return;
                }
                const { response } = await client.POST("/auth/sso/{provider}/link", {
                    params: {
                        path: {
                            provider: params.provider,
                        },
                    },
                    body: {
                        code: code,
                        state: state,
                    },
                });
                if (response.status == 200) {
                    window.opener.postMessage({status: LoginStatus.loginSuccess} as LoginMessage, window.location.origin);
                }else {
                    let msg: string
                    switch (response.status) {
                        case 400:
                            msg = "無法連結帳號，請重新登入。"
                            break
                        case 401:
                            msg = "請先登入。"
                            break
                        default:
                            msg = "連結失敗，請稍後再試。"
                            break
                    }
                    window.opener.postMessage({status: LoginStatus.loginFailed, error: msg} as LoginMessage, window.location.origin);
                }
            };
            await handleAuthCode();
        })();
    }, [props.params, router]);

    return (
        <div>登入處理中</div>
    );
}
