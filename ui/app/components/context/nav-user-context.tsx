'use client';

import { createContext, useContext, ReactNode, useEffect, useState } from 'react'
import { BACKEND_API_BASE_URL } from '@/app/constants'
import createClient from "openapi-fetch";
import type { paths } from "@/app/openapi/openapi";
import { usePathname } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { useToast } from '@/hooks/use-toast';
import { UserProfileDropdown } from '@/app/components/user-profile';
import { LoginPanel } from '@/app/components/login-panel';

type UserContextType = {
    username: string | null;
    isLoggedIn: boolean;
    openProfile: () => void;
    openLoginPanel: () => void;
    refreshUserProvider: () => void;
};

const UserContext = createContext<UserContextType | undefined>(undefined);
const client = createClient<paths>({ baseUrl: BACKEND_API_BASE_URL });

export function UserProvider({ children }: { children: ReactNode }) {
    const pathname = usePathname();
    const [componentStatus, setComponentStatus] = useState(false)
    const [isLoggedIn, setIsLoggedIn] = useState(false)
    const [username, setUsername] = useState<string | null>(null)
    const [isProfileOpen, setIsProfileOpen] = useState(false)
    const [isLoginPanelOpen, setIsLoginPanelOpen] = useState(false)
    const ctx: UserContextType = {
        username,
        isLoggedIn,
        openProfile: () => setIsProfileOpen(true),
        openLoginPanel: () => setIsLoginPanelOpen(true),
        refreshUserProvider: () => setComponentStatus(!componentStatus),
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
                setUsername(Buffer.from(username, 'base64').toString('utf-8'));
            }
            console.log('User status:', username);
        })();
    }, [pathname, componentStatus]);

    return (
        <UserContext.Provider value={ctx}>
            {children}
            {/* User Profile */}
            {isProfileOpen && <UserProfileDropdown onClose={() => setIsProfileOpen(false)} />}
            {isLoginPanelOpen && <LoginPanel onClose={() => setIsLoginPanelOpen(false)} />}
        </UserContext.Provider>
    );
}

export function LoginButton({
    children,
    ...props
}: React.ComponentProps<typeof Button>) {
    const { openLoginPanel } = useUser();
    return (
        <Button onClick={openLoginPanel} {...props}>{children}</Button>
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
