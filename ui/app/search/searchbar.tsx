'use client';

import { useRouter } from 'next/navigation'
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { X } from 'lucide-react'
import { Switch } from "@/components/ui/switch"

import type { paths, Defined } from "@/app/openapi"
import { useSearchBar } from './searchbar-context';
import { DateTimePicker } from '@/app/components/date-time-picker'

type searchRequestType = Defined<paths["/auction/items"]["get"]["parameters"]["query"]>

export function SearchBar({ searchRequest }: { searchRequest: searchRequestType }) {
    const router = useRouter()
    const { activeComponent, setActiveComponent } = useSearchBar();

    if (!activeComponent) return null;

    const handleSearch = (e: React.FormEvent<HTMLFormElement>) => {
        e.preventDefault()
        // Construct the query string from form
        const formData = new FormData(e.currentTarget)
        const filteredFormData = Array.from(formData)
            .filter((kv) => kv[1])
            .map(([k, v]) => [k, typeof v === 'string' ? v : v.name])
        const queryString = new URLSearchParams(filteredFormData).toString()

        // Navigate to the new URL
        if (queryString.length) {
            router.push(`/search?${queryString}`)
        } else {
            router.push(`/search`)
        }
        router.refresh()
    }

    return (
        <div className="md:w-1/3 md:max-w-[300px] transition-all duration-300 ease-in-out">
            <Card className="h-full bg-gradient-to-br from-primary/10 to-secondary/10 shadow-lg">
                <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
                    <CardTitle>搜尋選項</CardTitle>
                    <Button
                        variant="ghost"
                        size="icon"
                        onClick={() => setActiveComponent(false)}
                        aria-label="隱藏搜尋選項"
                    >
                        <X className="h-4 w-4" />
                    </Button>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSearch} className="space-y-4">
                        <div className="flex items-center space-x-2 mb-4">
                            <label htmlFor="excludeEnded" className="block text-sm font-medium mb-1">
                                排除已結束拍賣
                            </label>
                            <Switch id="excludeEnded" name="excludeEnded" defaultChecked={searchRequest.excludeEnded} />
                        </div>
                        <div>
                            <label htmlFor="name" className="block text-sm font-medium mb-1">
                                商品名稱
                            </label>
                            <Input id="title" name="title" placeholder="搜尋商品..." className="w-full" defaultValue={searchRequest.title} />
                        </div>
                        <div>
                            <label htmlFor="startPrice[from]" className="block text-sm font-medium mb-1">
                                起拍價格範圍
                            </label>
                            <div className="flex space-x-2">
                                <Input id="startPrice[from]" name="startPrice[from]" type="number" placeholder="最低" className="w-1/2" defaultValue={searchRequest.startPrice?.from} />
                                <Input name="startPrice[to]" type="number" placeholder="最高" className="w-1/2" defaultValue={searchRequest.startPrice?.to} />
                            </div>
                        </div>
                        <div>
                            <label htmlFor="currentBid[from]" className="block text-sm font-medium mb-1">
                                當前出價範圍
                            </label>
                            <div className="flex space-x-2">
                                <Input id="currentBid[from]" name="currentBid[from]" type="number" placeholder="最低" className="w-1/2" defaultValue={searchRequest.currentBid?.from} />
                                <Input name="currentBid[to]" type="number" placeholder="最高" className="w-1/2" defaultValue={searchRequest.currentBid?.to} />
                            </div>
                        </div>
                        <div>
                            <label htmlFor="startTime[from]" className="block text-sm font-medium mb-1">
                                拍賣開始時間範圍
                            </label>
                            <DateTimePicker name="startTime[from]" label="開始" className="w-1/2" defaultDate={searchRequest.startTime?.from} />
                            <DateTimePicker name="startTime[to]" label="結束" className="w-1/2" defaultDate={searchRequest.startTime?.to} />
                        </div>
                        <div>
                            <label htmlFor="endTime[from]" className="block text-sm font-medium mb-1">
                                拍賣結束時間範圍
                            </label>
                            <DateTimePicker name="endTime[from]" label="開始" className="w-1/2" defaultDate={searchRequest.endTime?.from} />
                            <DateTimePicker name="endTime[to]" label="結束" className="w-1/2" defaultDate={searchRequest.endTime?.to} />
                        </div>
                        <div>
                            <label htmlFor="sort[key]" className="block text-sm font-medium mb-1">
                                排序依據
                            </label>
                            <Select name="sort[key]" defaultValue={searchRequest.sort?.key}>
                                <SelectTrigger id="sort[key]">
                                    <SelectValue placeholder="選擇排序依據" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="name">名稱</SelectItem>
                                    <SelectItem value="startPrice">起拍價格</SelectItem>
                                    <SelectItem value="currentBid">當前出價</SelectItem>
                                    <SelectItem value="startTime">開始時間</SelectItem>
                                    <SelectItem value="endTime">結束時間</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <div>
                            <label htmlFor="sort[order]" className="block text-sm font-medium mb-1">
                                排序順序
                            </label>
                            <Select name="sort[order]" defaultValue={searchRequest.sort?.order}>
                                <SelectTrigger id="sort[order]">
                                    <SelectValue placeholder="選擇排序順序" />
                                </SelectTrigger>
                                <SelectContent>
                                    <SelectItem value="asc">升序</SelectItem>
                                    <SelectItem value="desc">降序</SelectItem>
                                </SelectContent>
                            </Select>
                        </div>
                        <Button type="submit" className="w-full">搜尋</Button>
                    </form>
                </CardContent>
            </Card>
        </div>
    )
}
