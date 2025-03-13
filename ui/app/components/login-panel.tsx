"use client"

import { useState, useEffect } from "react"
import { FaGoogle, FaGithub, FaMicrosoft } from "react-icons/fa"
import { SiAuthentik } from "react-icons/si"
import { FiX } from "react-icons/fi"
import { Button } from "@/components/ui/button"
import { useToast } from '@/hooks/use-toast';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { components, Defined } from "@/app/openapi";
import { login } from "@/app/components/user/login";
import { useUser } from '@/app/components/context/nav-user-context';

type SSOProviderType = Defined<components["schemas"]["SSOProvider"]>
const internalProvider: SSOProviderType = "internal"
const googleProvider: SSOProviderType = "google"
const githubProvider: SSOProviderType = "github"
const microsoftProvider: SSOProviderType = "microsoft"

interface LoginPanelProps {
    onClose: () => void
}

export function LoginPanel({ onClose }: LoginPanelProps) {
    const [isVisible, setIsVisible] = useState(false)
    const { toast } = useToast();
    const { refreshUserProvider: refresh } = useUser()
    
    // Animation effect on mount
    useEffect(() => {
        // Small delay to trigger animation
        const timer = setTimeout(() => setIsVisible(true), 50)
        return () => clearTimeout(timer)
    }, [])

    const handleClose = () => {
        setIsVisible(false)
        // Delay actual closing to allow animation to complete
        setTimeout(onClose, 300)
    }

    const handleLogin = async (provider: SSOProviderType) => {
        await login(provider, toast, () => {
            refresh()
            handleClose()
        })
    }

    return (
        <>
            {/* Backdrop overlay */}
            <div
                className={`fixed inset-0 bg-black transition-opacity duration-300 ease-in-out z-40 ${isVisible ? "opacity-50 dark:opacity-60" : "opacity-0"
                    }`}
                onClick={handleClose}
            />

            <Card
                className={`fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 
          w-[90%] max-w-md z-50 shadow-xl border-2 transition-all duration-300 ease-in-out
          ${isVisible ? "scale-100 opacity-100" : "scale-95 opacity-0"}`}
            >
                <CardHeader className="relative pb-2">
                    <Button variant="ghost" size="icon" className="absolute right-2 top-2" onClick={handleClose}>
                        <FiX className="h-4 w-4" />
                    </Button>
                    <CardTitle className="text-xl text-center">選擇登入方式</CardTitle>
                </CardHeader>
                <CardContent className="p-6">
                    <div className="space-y-4">

                        <Button
                            variant="outline"
                            className="w-full h-12 flex items-center justify-center gap-3 text-base hover:bg-muted/60"
                            onClick={() => handleLogin(internalProvider)}
                        >
                            <SiAuthentik className="h-5 w-5" />
                            <span>使用 Internal 登入</span>
                        </Button>

                        <Button
                            variant="outline"
                            className="w-full h-12 flex items-center justify-center gap-3 text-base hover:bg-muted/60"
                            onClick={() => handleLogin(googleProvider)}
                        >
                            <FaGoogle className="h-5 w-5" />
                            <span>使用 Google 登入</span>
                        </Button>

                        <Button
                            variant="outline"
                            className="w-full h-12 flex items-center justify-center gap-3 text-base hover:bg-muted/60"
                            disabled
                            onClick={() => handleLogin(githubProvider)}
                        >
                            <FaGithub className="h-5 w-5" />
                            <span>使用 Github 登入</span>
                        </Button>

                        <Button
                            variant="outline"
                            className="w-full h-12 flex items-center justify-center gap-3 text-base hover:bg-muted/60"
                            disabled
                            onClick={() => handleLogin(microsoftProvider)}
                        >
                            <FaMicrosoft className="h-5 w-5" />
                            <span>使用 Microsoft 登入</span>
                        </Button>
                    </div>

                    <div className="mt-6 text-center text-sm text-muted-foreground">
                        <p>登入即表示您同意我們不存在的</p>
                        <div className="mt-1 space-x-1">
                            <a className="underline hover:text-primary">
                                服務條款
                            </a>
                            <span>和</span>
                            <a className="underline hover:text-primary">
                                隱私政策
                            </a>
                        </div>
                    </div>
                </CardContent>
            </Card>
        </>
    )
}

