import Link from 'next/link'
import { Button } from "@/components/ui/button"
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import { Menu } from 'lucide-react'
import { NavLink, NavDropdownMenuLink } from '@/app/components/nav/nav-link'
import { NavDropdownMenuUser, NavUser } from '@/app/components/nav/nav-user'
import { UserProvider } from '@/app/components/context/nav-user-context'
import { ThemeTrigger } from '@/app/components/theme-provider'

export function Navbar() {
    return (
        <nav className="bg-background border-b">
            <div className="container mx-auto px-4">
                <div className="flex items-center justify-between h-16">
                    <UserProvider>
                        {/* Logo */}
                        <div className="flex items-center">
                            <Link href="/" className="text-2xl font-bold text-primary">
                                未定義拍賣網
                            </Link>
                        </div>
                        {/* Navbar */}
                        <div className="hidden md:block">
                            <div className="ml-10 flex items-baseline space-x-4">
                                <NavLink />
                            </div>
                        </div>
                        <div className="flex items-center space-x-2">
                            <ThemeTrigger />
                            {/* User */}
                            <div className="hidden md:block">
                                <NavUser />
                            </div>
                            {/* Mobile */}
                            <div className="md:hidden">
                                <DropdownMenu>
                                    <DropdownMenuTrigger asChild>
                                        <Button variant="outline" size="icon">
                                            <Menu className="h-5 w-5" />
                                            <span className="sr-only">打開選單</span>
                                        </Button>
                                    </DropdownMenuTrigger>
                                    <DropdownMenuContent align="end">
                                        <NavDropdownMenuLink />
                                        <NavDropdownMenuUser />
                                    </DropdownMenuContent>
                                </DropdownMenu>
                            </div>
                        </div>
                    </UserProvider>
                </div>
            </div>
        </nav>
    )
}

