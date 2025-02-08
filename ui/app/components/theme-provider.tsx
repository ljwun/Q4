'use client'

import * as React from "react"
import { Button } from "@/components/ui/button"
import { Moon, Sun } from 'lucide-react'
import { ThemeProvider as NextThemesProvider, useTheme } from "next-themes"

export function ThemeProvider({
    children,
    ...props
}: React.ComponentProps<typeof NextThemesProvider>) {
    return <NextThemesProvider {...props}>{children}</NextThemesProvider>
}

export function ThemeTrigger() {
    const { theme, setTheme } = useTheme()
    return (
        <Button
            variant="ghost"
            size="icon"
            onClick={() => setTheme(theme === "light" ? "dark" : "light")}
            aria-label="切換主題"
        >
            { theme === "light"? <Moon className="h-5 w-5" /> : <Sun className="h-5 w-5" /> }
        </Button>
    )
}
