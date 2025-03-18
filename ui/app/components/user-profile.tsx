"use client"

import { useState, useEffect } from "react"
import { FaGoogle, FaGithub, FaMicrosoft } from "react-icons/fa"
import { SiAuthentik } from "react-icons/si"
import { FiLogOut, FiX, FiLoader } from "react-icons/fi"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Card, CardContent, CardHeader, CardTitle, CardFooter } from "@/components/ui/card"
import { Separator } from "@/components/ui/separator"
import { LogoutButton } from '@/app/components/context/nav-user-context'
import { useToast } from '@/hooks/use-toast';
import createClient from "openapi-fetch";
import { BACKEND_API_BASE_URL } from '@/app/constants'
import { components, paths, Defined, SSOProvider } from "@/app/openapi";
import { useUser, LoginButton } from "@/app/components/context/nav-user-context"
import { LoginMessage, LoginStatus } from '@/app/components/user/login';

type SSOProviderType = Defined<components["schemas"]["SSOProvider"]>

interface UserProfileDropdownProps {
  onClose: () => void
}

interface UserData {
  username: string
  connectedAccounts: {
    internal: boolean
    google: boolean
    github: boolean
    microsoft: boolean
  }
}

export function UserProfileDropdown({ onClose }: UserProfileDropdownProps) {
  const [userData, setUserData] = useState<UserData>({ connectedAccounts: {} } as UserData)
  const [isEditing, setIsEditing] = useState(false)
  const [isVisible, setIsVisible] = useState(false)
  const [isLoading, setIsLoading] = useState(true)
  const [state, setState] = useState(false)
  const refresh = () => setState(!state)
  const { refreshUserProvider } = useUser()
  const { toast } = useToast();
  const client = createClient<paths>({ baseUrl: BACKEND_API_BASE_URL });

  // Animation effect on mount
  useEffect(() => {
    setIsLoading(true)
    // Small delay to trigger animation
    const timer = setTimeout(() => setIsVisible(true), 50)
    // 取得用戶資料
    const fetchUserData = async () => {
      const { data, error, response } = await client.GET("/user/info")
      if (response.status == 401) {
        toast({
          title: "載入用戶資料失敗",
          description: "請先登入以繼續。",
          variant: "destructive",
          action: <LoginButton>登入</LoginButton>
        });
        refreshUserProvider()
        handleClose()
        return
      } else if (!data || error || response.status !== 200) {
        toast({
          title: "載入用戶資料失敗",
          description: "無法取得用戶資料，請稍後再試。",
          variant: "destructive",
        });
        console.error('Error during fetching user data:', error);
        handleClose()
        return
      }
      setUserData({
        username: data.username,
        connectedAccounts: {
          internal: data.ssoProviders.Internal,
          google: data.ssoProviders.Google,
          github: data.ssoProviders.GitHub,
          microsoft: data.ssoProviders.Microsoft,
        },
      })
      setIsLoading(false)
    }

    fetchUserData()
    return () => clearTimeout(timer)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [state])

  const setUsername = (newUsername: string) => {
    setUserData((prev) => ({ ...prev, username: newUsername }))
  }

  const saveUsername = async () => {
    setIsEditing(false)
    setIsLoading(true)
    // 更新用戶名稱
    const { error, response } = await client.PATCH("/user/info", {
      body: { username: userData.username },
    });
    if (response.status == 401) {
      toast({
        title: "更新用戶名稱失敗",
        description: "請先登入以繼續。",
        variant: "destructive",
        action: <LoginButton>登入</LoginButton>
      });
      refreshUserProvider()
      handleClose()
      return
    } else if (response.status == 400) {
      toast({
        title: "更新用戶名稱失敗",
        description: "用戶名稱不符合規範，請重新輸入。",
        variant: "destructive",
      });
    } else if (error || response.status !== 200) {
      toast({
        title: "更新用戶名稱失敗",
        description: "無法更新用戶名稱，請稍後再試。",
        variant: "destructive",
      });
      console.error('Error during saving username:', error);
    } else {
      toast({
        title: "更新用戶名稱成功",
        description: "用戶名稱已更新。",
      });
    }
    refresh()
  }

  const disconnectSsoProvider = async (provider: SSOProviderType) => {
    console.log(`Disconnecting ${provider} account...`)
    setIsLoading(true)
    // 解除連結 SSO 帳號
    const { response } = await client.DELETE("/auth/sso/{provider}/link", {
      params: {
        path: { provider },
      },
    })
    if (response.status == 401) {
      toast({
        title: "解除連結失敗",
        description: "請先登入以繼續。",
        variant: "destructive",
        action: <LoginButton>登入</LoginButton>
      });
      refreshUserProvider()
      handleClose()
      return
    } else if (response.status == 404) {
      toast({
        title: "解除連結失敗",
        description: "無法解除連結，此SSO類型無法使用。",
        variant: "destructive",
      });
    } else if (response.status == 409) {
      toast({
        title: "解除連結失敗",
        description: "無法解除連結，需要留有至少一個連結的SSO帳號。",
        variant: "destructive",
      });
    } else if (response.status !== 200) {
      toast({
        title: "解除連結失敗",
        description: "無法解除連結，請稍後再試。",
        variant: "destructive",
      });
      console.error('Error during disconnecting SSO provider:', response.status);
    } else {
      toast({
        title: "解除連結成功",
        description: "帳號連結已解除。",
      });
    }
    refresh()
  }

  const connectSsoProvider = async (provider: SSOProviderType) => {
    console.log(`Connecting ${provider} account...`)
    setIsLoading(true)
    // 連結 SSO 帳號
    const { response } = await client.GET("/auth/sso/{provider}/login", {
      params: {
        path: { provider },
        query: {
          redirectUrl: new URL(`/sso/${provider}/link`, location.origin).toString(),
        },
      },
    });
    if (response.status == 404) {
      toast({
        title: "連結失敗",
        description: "目前不支援此SSO提供者。",
        variant: "destructive",
      });
      setIsLoading(false)
      return
    }
    const authUrl = response?.headers.get('Location')
    if (!authUrl || response.status != 200) {
      toast({
        title: "連結失敗",
        description: "無法取得到SSO登入網頁，請稍後再試。",
        variant: "destructive",
      });
      setIsLoading(false)
      return
    }
    const authWindow = window.open(authUrl, "authWindow");
    window.addEventListener("message", (event) => {
      if (event.origin !== location.origin) {
        return
      }
      // 檢查資料是否是LoginMessage類型
      if (!("status" in event.data)) {
        return
      }
      const message = event.data as LoginMessage
      switch (message.status) {
        case LoginStatus.loginSuccess:
          toast({
            title: "連結成功",
            description: "連結成功",
          });
          break
        case LoginStatus.loginFailed:
          toast({
            title: "連結失敗",
            description: message.error || "連結失敗，請稍後再試。",
            variant: "destructive",
          });
          break
      }
      refresh()
      authWindow?.close()
    })
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
        className={`fixed inset-0 bg-black transition-opacity duration-300 ease-in-out z-40 ${isVisible ? "opacity-50 dark:opacity-60" : "opacity-0"
          }`}
        onClick={isLoading ? undefined : handleClose}
      />

      <Card
        className={`fixed top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 
          w-[90%] max-w-md z-50 shadow-xl border-2 transition-all duration-300 ease-in-out
          ${isVisible ? "scale-100 opacity-100" : "scale-95 opacity-0"}
          ${isLoading ? "pointer-events-none" : ""}
        `}
      >
        {/* Loading overlay */}
        {isLoading && (
          <div className="absolute inset-0 bg-background/80 backdrop-blur-sm z-10 flex flex-col items-center justify-center rounded-lg">
            <div className="animate-spin mb-4">
              <FiLoader className="h-10 w-10 text-primary" />
            </div>
            <p className="text-center text-muted-foreground">載入用戶資料中...</p>
          </div>
        )}
        <CardHeader className="relative pb-2">
          <Button variant="ghost" size="icon" className="absolute right-2 top-2" onClick={handleClose} disabled={isLoading}>
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
              <Input
                id="username"
                value={userData.username}
                onChange={(e) => setUsername(e.target.value)}
                className="flex-1 h-9"
                autoFocus={isEditing}
                disabled={!isEditing}
              />
              {isEditing ? (
                <>
                  <Button size="sm" onClick={saveUsername} className="shrink-0">
                    保存
                  </Button>
                </>
              ) : (
                <>
                  <Button variant="outline" size="sm" onClick={() => setIsEditing(true)} className="shrink-0" disabled={isLoading}>
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
                  variant={userData.connectedAccounts.internal ? "destructive" : "outline"}
                  size="sm"
                  onClick={userData.connectedAccounts.internal ? () => disconnectSsoProvider(SSOProvider.Internal) : () => connectSsoProvider(SSOProvider.Internal)}
                >
                  {userData.connectedAccounts.internal ? "解除連結" : "連結"}
                </Button>
              </div>

              <div className="flex justify-between items-center bg-muted/50 p-3 rounded-lg">
                <div className="flex items-center space-x-3">
                  <FaGoogle className="h-5 w-5" />
                  <span className="text-sm font-medium">Google</span>
                </div>
                <Button
                  variant={userData.connectedAccounts.google ? "destructive" : "outline"}
                  size="sm"
                  onClick={userData.connectedAccounts.google ? () => disconnectSsoProvider(SSOProvider.Google) : () => connectSsoProvider(SSOProvider.Google)}
                >
                  {userData.connectedAccounts.google ? "解除連結" : "連結"}
                </Button>
              </div>

              <div className="flex justify-between items-center bg-muted/50 p-3 rounded-lg">
                <div className="flex items-center space-x-3">
                  <FaGithub className="h-5 w-5" />
                  <span className="text-sm font-medium">Github</span>
                </div>
                <Button
                  variant={userData.connectedAccounts.github ? "destructive" : "outline"}
                  size="sm"
                  onClick={userData.connectedAccounts.github ? () => disconnectSsoProvider(SSOProvider.GitHub) : () => connectSsoProvider(SSOProvider.GitHub)}
                >
                  {userData.connectedAccounts.github ? "解除連結" : "連結"}
                </Button>
              </div>

              <div className="flex justify-between items-center bg-muted/50 p-3 rounded-lg">
                <div className="flex items-center space-x-3">
                  <FaMicrosoft className="h-5 w-5" />
                  <span className="text-sm font-medium">Microsoft</span>
                </div>
                <Button
                  variant={userData.connectedAccounts.microsoft ? "destructive" : "outline"}
                  size="sm"
                  onClick={userData.connectedAccounts.microsoft ? () => disconnectSsoProvider(SSOProvider.Microsoft) : () => connectSsoProvider(SSOProvider.Microsoft)}
                >
                  {userData.connectedAccounts.microsoft ? "解除連結" : "連結"}
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


