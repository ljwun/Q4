import '@/app/globals.css';
import type { Metadata } from "next";
import { Inter } from 'next/font/google'
import { Navbar } from '@/app/components/navbar'
import { Toaster } from "@/components/ui/toaster"
import { ThemeProvider } from "@/app/components/theme-provider"
import { PublicEnvScript } from 'next-runtime-env';

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: '未定義拍賣網',
  description: '發現並競拍獨特的未定義商品',
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="zh-TW" suppressHydrationWarning>
      <head>
        <PublicEnvScript />
      </head>
      <body
        className={inter.className}
      >
        <ThemeProvider
          attribute="class"
          defaultTheme="system"
          enableSystem
          disableTransitionOnChange
        >
          <Navbar />
          <main>{children}</main>
          <Toaster />
        </ThemeProvider>
      </body>
    </html>
  );
}
