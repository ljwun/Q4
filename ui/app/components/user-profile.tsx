"use client"

import { useState, useEffect } from "react"
import { FaGoogle, FaGithub, FaMicrosoft } from "react-icons/fa"
import { SiAuthentik } from "react-icons/si"
import { FiLogOut, FiX } from "react-icons/fi"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { LogoutButton } from '@/app/components/context/nav-user-context'

interface UserProfileDropdownProps {
  onClose: () => void
}

export function UserProfileDropdown({ onClose }: UserProfileDropdownProps) {
  const [username, setUsername] = useState("用戶名稱")
  const [isEditing, setIsEditing] = useState(false)
  const [isVisible, setIsVisible] = useState(false)

  // Mock connected accounts
  const [connectedAccounts] = useState({
    internal: false,
    google: true,
    github: false,
    microsoft: false,
  })

  // Animation effect on mount
  useEffect(() => {
    // Small delay to trigger animation
    const timer = setTimeout(() => setIsVisible(true), 50)
    return () => clearTimeout(timer)
  }, [])

  const saveUsername = () => {
    alert('該功能尚未實現')
    setIsEditing(false)
    // Here you would typically save the username to your backend
  }

  const handleClose = () => {
    setIsVisible(false)
    // Delay actual closing to allow animation to complete
    setTimeout(onClose, 300)
  }

  return (
    <>
      {/* Backdrop overlay */}
      <div
        className={`fixed inset-0 bg-black transition-opacity duration-300 ease-in-out z-40 ${
          isVisible ? "opacity-50 dark:opacity-60" : "opacity-0"
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
          <CardTitle className="text-xl">用戶設定</CardTitle>
        </CardHeader>
        <CardContent className="space-y-6">
          {/* Username section */}
          <div className="space-y-3 bg-muted/50 p-3 rounded-lg">
            <div className="flex items-center gap-2">
              <Label htmlFor="username" className="w-20 shrink-0 text-sm font-medium">
                用戶名稱
              </Label>
              {isEditing ? (
                <>
                  <Input
                    id="username"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                    className="flex-1 h-9"
                    autoFocus
                  />
                  <Button size="sm" onClick={saveUsername} className="shrink-0">
                    保存
                  </Button>
                </>
              ) : (
                <>
                  <span className="flex-1 text-sm truncate">{username}</span>
                  <Button variant="outline" size="sm" onClick={() => setIsEditing(true)} className="shrink-0">
                    編輯
                  </Button>
                </>
              )}
            </div>
          </div>

          <Separator />

          {/* SSO accounts section */}
          <div className="space-y-4">
            <Label className="text-base">連結的帳號</Label>

            <div className="space-y-4">
              <div className="flex justify-between items-center bg-muted/50 p-3 rounded-lg">
                <div className="flex items-center space-x-3">
                  <SiAuthentik className="h-5 w-5" />
                  <span className="text-sm font-medium">Internal</span>
                </div>
                <Button
                  variant={connectedAccounts.internal ? "destructive" : "outline"}
                  size="sm"
                  disabled
                >
                  {connectedAccounts.internal ? "解除連結" : "連結"}
                </Button>
              </div>

              <div className="flex justify-between items-center bg-muted/50 p-3 rounded-lg">
                <div className="flex items-center space-x-3">
                  <FaGoogle className="h-5 w-5" />
                  <span className="text-sm font-medium">Google</span>
                </div>
                <Button
                  variant={connectedAccounts.google ? "destructive" : "outline"}
                  size="sm"
                  disabled
                >
                  {connectedAccounts.google ? "解除連結" : "連結"}
                </Button>
              </div>

              <div className="flex justify-between items-center bg-muted/50 p-3 rounded-lg">
                <div className="flex items-center space-x-3">
                  <FaGithub className="h-5 w-5" />
                  <span className="text-sm font-medium">Github</span>
                </div>
                <Button
                  variant={connectedAccounts.github ? "destructive" : "outline"}
                  size="sm"
                  disabled
                >
                  {connectedAccounts.github ? "解除連結" : "連結"}
                </Button>
              </div>

              <div className="flex justify-between items-center bg-muted/50 p-3 rounded-lg">
                <div className="flex items-center space-x-3">
                  <FaMicrosoft className="h-5 w-5" />
                  <span className="text-sm font-medium">Microsoft</span>
                </div>
                <Button
                  variant={connectedAccounts.microsoft ? "destructive" : "outline"}
                  size="sm"
                  disabled
                >
                  {connectedAccounts.microsoft ? "解除連結" : "連結"}
                </Button>
              </div>
            </div>
          </div>
        </CardContent>
        <CardFooter className="flex justify-end pt-2 pb-4">
          <LogoutButton variant="destructive">
            <FiLogOut className="h-4 w-4" />
            登出
          </LogoutButton>
        </CardFooter>
      </Card>
    </>
  )
}


