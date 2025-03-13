'use client'

import Link from 'next/link'
import { Button } from "@/components/ui/button"
import { DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { useUser, LoginButton, LogoutButton } from '@/app/components/context/nav-user-context'
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import { FiUser } from 'react-icons/fi'
import { getAvatarUrl } from '@/app/components/user/avatar'
import { useTheme } from "next-themes"

export function NavUser() {
    const { isLoggedIn, username, openProfile } = useUser()
    const { theme } = useTheme()
    if (!isLoggedIn || username == null) {
        return (
            <>
                <LoginButton variant="outline" className="mr-2">登入</LoginButton>
                <Button>
                    <Link href="/register">註冊</Link>
                </Button>
            </>
        )
    }
    return (
        <Button
            variant="ghost"
            size="icon"
            className="rounded-full mr-2"
            onClick={openProfile}
        >
            <Avatar>
                <AvatarImage src={getAvatarUrl(username, theme == "dark")} />
                <AvatarFallback>
                    <FiUser className="h-5 w-5" />
                </AvatarFallback>
            </Avatar>
        </Button>
    )
}

export function NavDropdownMenuUser() {
    const { isLoggedIn, username, openProfile } = useUser()
    if (!isLoggedIn) {
        return (
            <>
                <DropdownMenuItem asChild>
                    <LoginButton>登入</LoginButton>
                </DropdownMenuItem>
                <DropdownMenuItem asChild>
                    <Link href="/register">註冊</Link>
                </DropdownMenuItem>
            </>
        )
    }
    return (
        <>
            <DropdownMenuItem asChild>
                <Button variant="outline" className="mr-2" onClick={openProfile}>{username}</Button>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
                <LogoutButton>登出</LogoutButton>
            </DropdownMenuItem>
        </>
    )
}