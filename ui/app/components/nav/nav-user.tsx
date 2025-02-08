'use client'

import Link from 'next/link'
import { Button } from "@/components/ui/button"
import { DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { useUser, LoginButton, LogoutButton } from '@/app/components/context/nav-user-context'

export function NavUser() {
    const { isLoggedIn, username } = useUser()
    if (!isLoggedIn) {
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
        <>
            <span className="text-foreground mr-3" >{username}</span>
            <LogoutButton variant="outline">登出</LogoutButton>
        </>
    )
}

export function NavDropdownMenuUser() {
    const { isLoggedIn, username } = useUser()
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
                <span className="text-foreground mr-2" >{username}</span>
            </DropdownMenuItem>
            <DropdownMenuItem asChild>
                <LogoutButton>登出</LogoutButton>
            </DropdownMenuItem>
        </>
    )
}