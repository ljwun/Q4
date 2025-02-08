import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { ScrollArea } from "@/components/ui/scroll-area"
import type { components, Defined } from "@/app/openapi"

type BidEvent = Defined<components["schemas"]["BidEvent"]>

export function BidHistory({ bidRecords }: { bidRecords: BidEvent[] }) {
    return (
        <Dialog>
            <DialogTrigger asChild>
                <Button variant="outline">查看出價記錄</Button>
            </DialogTrigger>
            <DialogContent className="max-w-md">
                <DialogHeader>
                    <DialogTitle>出價記錄</DialogTitle>
                </DialogHeader>
                <ScrollArea className="h-[60vh]">
                    <ul className="space-y-2">
                        {bidRecords.length > 0 ? (
                            bidRecords.map((bid, index) => (
                                <li key={index} className="text-sm border-b pb-2">
                                    <span className="font-semibold">{bid.user || 'Unknown'}</span> - ${bid.bid || 0}
                                    <br />
                                    <span className="text-xs text-muted-foreground">{bid.time.toLocaleString()}</span>
                                </li>
                            ))
                        ) : (
                            <li className="text-sm">暫無出價記錄</li>
                        )}
                    </ul>
                </ScrollArea>
            </DialogContent>
        </Dialog>
    )
}

