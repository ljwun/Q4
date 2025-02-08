'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { DropdownMenuItem } from "@/components/ui/dropdown-menu"
import { useUser } from '@/app/components/context/nav-user-context'

const navItems = [
    { href: '/', label: '首頁', shouldLogin: false },
    { href: '/search', label: '搜尋', shouldLogin: false },
    { href: '/create-auction', label: '建立拍賣', shouldLogin: true },
    { href: '/about', label: '關於我們', shouldLogin: false },
]

export function NavLink() {
    const pathname = usePathname()
    const {isLoggedIn} = useUser()

    return (
        // 過濾出navItems中可顯示的選項
        navItems
            .filter((item) => !item.shouldLogin || isLoggedIn)
            .map((item) => (
                <Link
                    key={item.href}
                    href={item.href}
                    className={`px-3 py-2 rounded-md text-sm font-medium ${pathname === item.href
                        ? 'bg-primary text-primary-foreground'
                        : 'text-foreground hover:bg-accent hover:text-accent-foreground'
                        }`}
                >
                    {item.label}
                </Link>
            ))
    )
}

export function NavDropdownMenuLink() {
    const {isLoggedIn} = useUser()

    return (
        // 過濾出navItems中可顯示的選項
        navItems
            .filter((item) => !item.shouldLogin || isLoggedIn)
            .map((item) => (
                <DropdownMenuItem key={item.href} asChild>
                    <Link href={item.href}>{item.label}</Link>
                </DropdownMenuItem>
            ))
    )
}