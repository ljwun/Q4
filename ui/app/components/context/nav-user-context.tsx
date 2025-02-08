'use client';

import { createContext, useContext, ReactNode, useEffect, useState } from 'react'
import { BACKEND_API_BASE_URL } from '@/app/constants'
import createClient from "openapi-fetch";
import type { paths } from "@/app/openapi/openapi";
import { usePathname } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { makeCallbackUrl } from '@/app/constants';
import { useToast } from '@/hooks/use-toast';

type UserContextType = {
    username: string | null;
    isLoggedIn: boolean;
};

const UserContext = createContext<UserContextType | undefined>(undefined);
const client = createClient<paths>({ baseUrl: BACKEND_API_BASE_URL });

export function UserProvider({ children }: { children: ReactNode }) {
    const pathname = usePathname();
    const [isLoggedIn, setIsLoggedIn] = useState(false)
    const [username, setUsername] = useState<string | null>(null)
    const ctx: UserContextType = {
        username,
        isLoggedIn,
    };

    useEffect(() => {
        (async () => {
            // 檢查是否登入
            const username = getCookie('username')
            if (!username) {
                setIsLoggedIn(false);
                setUsername(null);
            } else {
                setIsLoggedIn(true);
                setUsername(username);
            }
        })();
    }, [pathname]);

    return (
        <UserContext.Provider value={ctx}>
            {children}
        </UserContext.Provider>
    );
}

export function LoginButton({
    children,
    ...props
}: React.ComponentProps<typeof Button>) {
    const { toast } = useToast();
    async function handleLogin() {
        const { error, response } = await client.GET("/auth/login", {
            params: {
                query: {
                    redirect_url: makeCallbackUrl(new URL(location.href)),
                },
            },
        });
        if (!error) {
            const authUrl = response?.headers.get('Location')
            if (authUrl) {
                location.href = authUrl;
                return
            }else{
                console.error('Error during login: Authorization URL not received');
            }
        }else{
            console.error('Error during login:', error);
        }
        toast({
            title: "登入失敗",
            description: "無法跳轉到登入網頁，請稍後再試。",
            variant: "destructive",
        });
    }
    return (
        <Button onClick={handleLogin} {...props}>{children}</Button>
    );
}

export function LogoutButton({
    children,
    ...props
}: React.ComponentProps<typeof Button>) {
    const { toast } = useToast();
    async function handleLogout() {
        const { error } = await client.GET("/auth/logout");
        if (error) {
            toast({
                title: "登出失敗",
                description: "無法登出，請稍後再試。",
                variant: "destructive",
            });
            console.error('Error during logout:', error);
            return;
        }
        location.reload();
    }
    return (
        <Button onClick={handleLogout} {...props}>{children}</Button>
    );
}

export function useUser() {
    const context = useContext(UserContext);
    if (!context) {
        throw new Error('useUser must be used within a UserProvider');
    }
    return context;
}

const getCookie = (name: string) => {
    return document.cookie.split('; ').reduce((r, v) => {
        const parts = v.split('=');
        return parts[0] === name ? decodeURIComponent(parts[1]) : r;
    }, '');
};
