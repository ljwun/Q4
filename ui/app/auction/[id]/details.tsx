import { Card, CardContent } from "@/components/ui/card"
import type { paths, Defined } from "@/app/openapi"

type ItemInfo = Defined<paths["/auction/item/{itemID}"]["get"]["responses"]["200"]["content"]["application/json"]>

export function AuctionDetails({ info }: { info: ItemInfo }) {
    return (
        <Card className="bg-gradient-to-br from-primary/10 to-secondary/10 shadow-lg">
            <CardContent className="p-6">
                <h3 className="text-xl font-semibold mb-4">拍賣信息</h3>
                <div className="grid grid-cols-2 gap-4">
                    <div>
                        <p className="text-sm text-muted-foreground">起拍價</p>
                        <p className="font-semibold">${info.startPrice ?? 'N/A'}</p>
                    </div>
                    <div>
                        <p className="text-sm text-muted-foreground">起拍時間</p>
                        <p className="font-semibold">{info.startTime.toLocaleString()}</p>
                    </div>
                    <div>
                        <p className="text-sm text-muted-foreground">保證金</p>
                        <p className="font-semibold">$0</p>
                    </div>
                    <div>
                        <p className="text-sm text-muted-foreground">競拍增幅</p>
                        <p className="font-semibold">$1</p>
                    </div>
                </div>
            </CardContent>
        </Card>
    )
}

