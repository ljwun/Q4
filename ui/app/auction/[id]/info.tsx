'use client'

import type { paths, components, Defined } from "@/app/openapi"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent } from "@/components/ui/card"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import { AuctionTimer } from "@/app/auction/[id]/auction-timer"
import {
    Carousel,
    CarouselContent,
    CarouselItem,
    CarouselNext,
    CarouselPrevious,
} from "@/components/ui/carousel"
import Image from 'next/image'
import { useCallback, useEffect, useState } from "react"
import { BidHistory } from "@/app/auction/[id]/history"
import { AuctionDetails } from "@/app/auction/[id]/details"
import { BACKEND_API_BASE_URL } from '@/app/constants'
import { dateReviver } from "@/app/utils"
import createClient from "openapi-fetch"
import { useToast } from "@/hooks/use-toast"
import { LoginButton } from "@/app/components/context/nav-user-context"

enum AuctionState {
    NOT_STARTED,
    IN_PROGRESS,
    ENDED
}

type ItemInfo = Defined<paths["/auction/item/{itemID}"]["get"]["responses"]["200"]["content"]["application/json"]>
type BidEvent = Defined<components["schemas"]["BidEvent"]>


export function AuctionItemInfo({ id, info }: { id: string, info: ItemInfo }) {
    const [userBid, setUserBid] = useState<number>(0)
    const [bidRecords, setBidRecords] = useState<BidEvent[]>(info.bidRecords)
    const [currentBid, setCurrentBid] = useState<BidEvent | undefined>(info.bidRecords[0])
    const [auctionState, setAuctionState] = useState<AuctionState>(AuctionState.NOT_STARTED)
    const [eventSource, setEventSource] = useState<EventSource>()
    const { toast } = useToast()
    const client = createClient<paths>({ baseUrl: BACKEND_API_BASE_URL })

    const startServerSideEventConnection = useCallback(() => {
        if (eventSource) {
            return
        }
        console.log('Starting server-sent events connection')
        try {
            const source = new EventSource(BACKEND_API_BASE_URL + `/auction/item/${id}/events`)
            source.addEventListener('bid', (event) => {
                const newBid = JSON.parse(event.data, dateReviver) as BidEvent
                setCurrentBid(newBid)
                setBidRecords((prev) => [newBid, ...prev])
            })
            setEventSource(source)
        }catch (e) {
            console.error('Failed to start server-sent events connection:', e)
        }
    }, [eventSource, id])

    const stopServerSideEventConnection = useCallback(() => {
        if (eventSource) {
            console.log('Closing server-sent events connection')
            eventSource.close()
            setEventSource(undefined)
        }
    }, [eventSource])

    useEffect(() => {
        const updateAuctionState = () => {
            const now = new Date().getTime()
            const startTime = info.startTime.getTime() || 0
            const endTime = info.endTime.getTime() || 0

            if (now < startTime) {
                setAuctionState(AuctionState.NOT_STARTED)
                if (now >= startTime - 60000) {
                    startServerSideEventConnection()
                }
            } else if (now >= startTime && now < endTime) {
                setAuctionState(AuctionState.IN_PROGRESS)
                startServerSideEventConnection()
            } else {
                setAuctionState(AuctionState.ENDED)
                stopServerSideEventConnection()
            }
        }

        // Initial state update
        updateAuctionState()

        // Update state every second
        const interval = setInterval(updateAuctionState, 1000)

        return () => {
            clearInterval(interval)
            stopServerSideEventConnection()
        }
    }, [info.startTime, info.endTime, startServerSideEventConnection, stopServerSideEventConnection])

    async function handleBid() {
        const { error, response } = await client.POST("/auction/item/{itemID}/bids", {
            params: {
                path: {
                    itemID: id,
                },
            },
            body: {
                bid: userBid,
            },
        })
        switch (response.status) {
            case 201:
                toast({
                    title: '出價成功',
                    description: '您的出價已提交',
                });
                setUserBid(0)
                break
            case 400:
                toast({
                    title: '出價失敗',
                    description: '出價金額太低',
                    variant: 'destructive'
                });
                break
            case 401:
                toast({
                    title: '出價失敗',
                    description: '請先登入',
                    variant: 'destructive',
                    action: <LoginButton>登入</LoginButton>
                });
                break
            case 403:
                toast({
                    title: '出價失敗',
                    description: '拍賣尚未開始',
                    variant: 'destructive'
                });
                break
            case 404:
                toast({
                    title: '出價失敗',
                    description: '拍賣不存在',
                    variant: 'destructive'
                });
                break
            case 410:
                toast({
                    title: '出價失敗',
                    description: '拍賣已結束',
                    variant: 'destructive'
                });
                break
            case 200:
                toast({
                    title: '出價成功',
                    description: '您已是最高出價者',
                });
                break
            default:
                toast({
                    title: '出價失敗',
                    description: '請稍後再試',
                    variant: 'destructive'
                });
                console.error('Failed to bid:', error);
                break
        }
    }

    return (
        <div className="max-w-6xl mx-auto">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
                <div className="space-y-6">
                    <Carousel className="w-full max-w-xl mx-auto">
                        <CarouselContent>
                            {[...Array(5)].map((_, index) => (
                                <CarouselItem key={index}>
                                    <div className="aspect-square">
                                        <Image
                                            src={`/placeholder-400x400.webp`}
                                            alt={`商品圖片 ${index + 1}`}
                                            width={400}
                                            height={400}
                                            className="w-full rounded-lg object-cover"
                                        />
                                    </div>
                                </CarouselItem>
                            ))}
                        </CarouselContent>
                        <CarouselPrevious />
                        <CarouselNext />
                    </Carousel>
                </div>
                <div className="space-y-6">
                    <Card className="bg-gradient-to-br from-primary/10 to-secondary/10 shadow-lg">
                        <CardContent className="p-6">
                            <h2 className="text-3xl font-bold mb-4">{info.title || 'Unnamed Item'}</h2>
                            <div className="text-2xl font-semibold mb-2">當前出價：${currentBid?.bid || info.startPrice}</div>
                            <div className="flex items-center justify-between mb-4">
                                <span className="text-lg">
                                    {currentBid ? `出價人：${currentBid.user}` : "暫無出價"}
                                </span>
                                <BidHistory bidRecords={bidRecords} />
                            </div>
                            {auctionState !== AuctionState.ENDED && (
                                <AuctionTimer 
                                    endTime={auctionState === AuctionState.NOT_STARTED 
                                        ? info.startTime.getTime() || 0 
                                        : info.endTime.getTime() || 0
                                    }
                                    mode={auctionState === AuctionState.NOT_STARTED ? 'start' : 'end'}
                                />
                            )}
                            <div className="flex space-x-2 mt-4">
                                <Input
                                    type="number"
                                    placeholder="您的出價"
                                    className="flex-grow"
                                    value={userBid || ''}
                                    onChange={(e) => setUserBid(Number(e.target.value))}
                                    disabled={auctionState !== AuctionState.IN_PROGRESS}
                                />
                                <Button 
                                    size="lg" 
                                    onClick={handleBid}
                                    disabled={auctionState !== AuctionState.IN_PROGRESS}
                                >
                                    {auctionState === AuctionState.NOT_STARTED 
                                        ? "拍賣尚未開始" 
                                        : auctionState === AuctionState.ENDED 
                                            ? "拍賣已結束" 
                                            : "出價"
                                    }
                                </Button>
                            </div>
                        </CardContent>
                    </Card>
                    <AuctionDetails info={info} />
                </div>
            </div>
            <Tabs defaultValue="description" className="mt-6">
                <TabsList className="grid w-full grid-cols-3">
                    <TabsTrigger value="description">詳細說明</TabsTrigger>
                    <TabsTrigger value="shipping">運輸信息</TabsTrigger>
                    <TabsTrigger value="terms">拍賣條款</TabsTrigger>
                </TabsList>
                <TabsContent value="description" className="mt-4">
                    <Card>
                        <CardContent className="p-4">
                            <h3 className="font-semibold mb-2">商品詳細說明</h3>
                            <div dangerouslySetInnerHTML={{ __html: info.description || '暫無描述' }}></div>
                        </CardContent>
                    </Card>
                </TabsContent>
                <TabsContent value="shipping" className="mt-4">
                    <Card>
                        <CardContent className="p-4">
                            <h3 className="font-semibold mb-2">運輸信息</h3>
                            <p>運費由買家承擔，預計費用為$200-$500，具體視運輸距離而定。商品將由專業人員包裝，全程保險，預計在拍賣結束後7-14個工作日內發貨。</p>
                        </CardContent>
                    </Card>
                </TabsContent>
                <TabsContent value="terms" className="mt-4">
                    <Card>
                        <CardContent className="p-4">
                            <h3 className="font-semibold mb-2">拍賣條款</h3>
                            <ul className="list-disc list-inside space-y-2">
                                以下皆為自動生成:
                                <li>所有競標均為最終出價，不可撤回。</li>
                                <li>買家需在拍賣結束後3個工作日內完成付款。</li>
                                <li>如買家未能及時付款，保證金將被沒收，商品將給予下一位最高出價者。</li>
                                <li>賣家保證商品描述真實，如實際商品與描述不符，買家可在收到商品後3天內申請退貨。</li>
                                <li>拍賣平台收取最終成交價的10%作為服務費。</li>
                            </ul>
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    )
}
