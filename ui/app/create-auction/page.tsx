'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardFooter, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Label } from "@/components/ui/label"
import { useToast } from "@/hooks/use-toast"
import { useEditor, EditorContent } from '@tiptap/react'
import StarterKit from '@tiptap/starter-kit'
import Underline from '@tiptap/extension-underline'
import Image from '@tiptap/extension-image'
import { Toggle } from "@/components/ui/toggle"
import { Bold, Italic, UnderlineIcon, Heading1, Heading2, Heading3, ImageIcon } from 'lucide-react'
import { BACKEND_API_BASE_URL } from '@/app/constants'
import createClient from "openapi-fetch";
import type { paths } from "@/app/openapi/openapi"
import { FileHandler } from '@/app/create-auction/file-handler'
import { DateTimePicker } from "@/app/components/date-time-picker"
import { LoginButton } from "@/app/components/context/nav-user-context"

export default function CreateAuctionPage() {
  const router = useRouter()
  const { toast } = useToast()
  const [isLoading, setIsLoading] = useState(false)
  const client = createClient<paths>({ baseUrl: BACKEND_API_BASE_URL });

  const handleImageUpload = async (file: File): Promise<string | null> => {
    const { error, response } = await client.POST("/image", {
      body: file,
      // 注: 預設的bodySerializer會將body處理成JSON格式，所以像是form或Blob等非JSON格式的body需要自行處理
      // 參考: https://github.com/openapi-ts/openapi-typescript/issues/1214
      bodySerializer: () => file,
      headers: {
        'Content-Type': file.type,
      },
    });
    switch (response.status) {
      case 201:
        toast({
          title: "上傳成功",
          description: "圖片上傳成功。",
        });
        return response.headers.get('Location')
      case 400:
        toast({
          title: "上傳失敗",
          description: "無法上傳圖片，圖片不符合上傳條件。",
          variant: "destructive",
        });
        break
      case 401:
        toast({
          title: '上傳失敗',
          description: '請先登入',
          variant: 'destructive',
          action: <LoginButton>登入</LoginButton>
        });
        break
      case 429:
        toast({
          title: '上傳失敗',
          description: '上傳次數過多，請過段時間後再上傳',
          variant: 'destructive'
        });
        break
      default:
        toast({
          title: '上傳失敗',
          description: '無法上傳圖片，請稍後再試',
          variant: 'destructive'
        });
        console.error('Failed to bid:', error);
        break
    }
    return null;
  }
  const editor = useEditor({
    extensions: [
      StarterKit,
      Underline,
      Image,
      FileHandler.configure({
        onUpload: handleImageUpload,
      }),
    ],
    content: '<p>描述您的拍賣項目...</p>',
    editorProps: {
      attributes: {
        class: 'min-h-[200px] prose prose-sm sm:prose lg:prose-lg xl:prose-2xl focus:outline-none',
      },
    },
  })

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    setIsLoading(true)

    // 獲取所有圖片 URI
    const imageUris = editor?.getJSON().content
      ?.filter(node => node.type === 'image')
      .map(node => node.attrs?.src)
      .filter((uri, index, self) => uri && self.indexOf(uri) === index) || [];

    const formData = new FormData(event.currentTarget)
    const { error, response } = await client.POST("/auction/item", {
      body: {
        title: formData.get('title') as string,
        startingPrice: parseInt(formData.get('startingPrice') as string, 10),
        startTime: new Date(formData.get('startTime') as string),
        endTime: new Date(formData.get('endTime') as string),
        description: editor?.getHTML(),
        carousels: imageUris,
      },
    });
    setIsLoading(false)
    if (error || response.status !== 201) {
      console.error('Failed to create auction:', error);
      toast({
        title: "創建失敗",
        description: "無法創建拍賣項目，請稍後再試。",
        variant: "destructive",
      });
      return;
    }
    toast({
      title: "拍賣創建成功",
      description: "您的拍賣項目已成功發布。",
    })
    const itemID = response.headers.get('Location') as string;
    router.push(`/auction/${itemID}`);
  }

  const addImage = () => {
    const input = document.createElement('input')
    input.type = 'file'
    input.accept = 'image/*'
    input.onchange = async () => {
      if (input.files?.length) {
        const file = input.files[0]
        const imageUrl = await handleImageUpload(file)
        if (imageUrl) {
          editor?.chain().focus().setImage({ src: imageUrl }).run()
        }
      }
    }
    input.click()
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <Card className="max-w-2xl mx-auto">
        <CardHeader>
          <CardTitle>創建新拍賣</CardTitle>
          <CardDescription>填寫以下信息來創建您的拍賣項目</CardDescription>
        </CardHeader>
        <form onSubmit={handleSubmit}>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="title">拍賣標題</Label>
              <Input id="title" name="title" placeholder="輸入拍賣項目的標題" required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="startingPrice">起拍價格</Label>
              <Input
                id="startingPrice"
                name="startingPrice"
                type="number"
                inputMode="numeric"
                placeholder="0"
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="startTime">開始時間</Label>
              <DateTimePicker name="startTime" defaultDate={new Date()} required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="endTime">結束時間</Label>
              <DateTimePicker name="endTime" defaultDate={new Date(new Date().getTime() + 5 * 3600000)} required />
            </div>
            <div className="space-y-2">
              <Label htmlFor="description">描述</Label>
              <div className="border rounded-md p-2">
                <div className="flex items-center space-x-2 mb-2">
                  <Toggle
                    pressed={editor?.isActive('bold') ?? false}
                    onPressedChange={() => editor?.chain().focus().toggleBold().run()}
                    disabled={!editor}
                  >
                    <Bold className="h-4 w-4" />
                  </Toggle>
                  <Toggle
                    pressed={editor?.isActive('italic') ?? false}
                    onPressedChange={() => editor?.chain().focus().toggleItalic().run()}
                    disabled={!editor}
                  >
                    <Italic className="h-4 w-4" />
                  </Toggle>
                  <Toggle
                    pressed={editor?.isActive('underline') ?? false}
                    onPressedChange={() => editor?.chain().focus().toggleUnderline().run()}
                    disabled={!editor}
                  >
                    <UnderlineIcon className="h-4 w-4" />
                  </Toggle>
                  <Toggle
                    pressed={editor?.isActive('heading', { level: 1 }) ?? false}
                    onPressedChange={() => editor?.chain().focus().toggleHeading({ level: 1 }).run()}
                    disabled={!editor}
                  >
                    <Heading1 className="h-4 w-4" />
                  </Toggle>
                  <Toggle
                    pressed={editor?.isActive('heading', { level: 2 }) ?? false}
                    onPressedChange={() => editor?.chain().focus().toggleHeading({ level: 2 }).run()}
                    disabled={!editor}
                  >
                    <Heading2 className="h-4 w-4" />
                  </Toggle>
                  <Toggle
                    pressed={editor?.isActive('heading', { level: 3 }) ?? false}
                    onPressedChange={() => editor?.chain().focus().toggleHeading({ level: 3 }).run()}
                    disabled={!editor}
                  >
                    <Heading3 className="h-4 w-4" />
                  </Toggle>
                  <Button
                    type="button"
                    variant="outline"
                    size="icon"
                    onClick={addImage}
                    disabled={!editor}
                  >
                    <ImageIcon className="h-4 w-4" />
                  </Button>
                </div>
                <EditorContent editor={editor} />
              </div>
            </div>
          </CardContent>
          <CardFooter>
            <Button type="submit" className="w-full" disabled={isLoading}>
              {isLoading ? '創建中...' : '創建拍賣'}
            </Button>
          </CardFooter>
        </form>
      </Card>
    </div>
  )
}

