import Link from 'next/link';
import Image from 'next/image';
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import type { Defined, paths } from "@/app/openapi";

type AuctionItemData = Defined<
	paths["/auction/items"]["get"]["responses"]["200"]["content"]["application/json"]["items"]
>[number];

export function AuctionItem({ item }: { item: AuctionItemData }) {
	return (
		<Link href={`/auction/${item.id}`}>
			<Card className={`h-full hover:shadow-lg transition-shadow shadow-lg relative ${
				item.isEnded ? "bg-gray-200 dark:bg-gray-800" : "bg-gradient-to-br from-primary/10 to-secondary/10 dark:from-primary/20 dark:to-secondary/20"
			}`}>
				<CardHeader>
					<CardTitle className='truncate'>{item.title}</CardTitle>
				</CardHeader>
				{item.isEnded && (
					<div className="absolute inset-0 bg-gray-500/30 dark:bg-gray-700/50 backdrop-blur-[1px] flex flex-col items-center justify-center z-10">
						<div className="bg-red-600 text-white px-4 py-2 rounded-full font-bold transform -rotate-12">
							已結標
						</div>
						<div className="mt-2 bg-white/90 dark:bg-gray-800/90 px-3 py-1 rounded-lg text-sm">
							結標價格：${item.currentBid}
						</div>
					</div>
				)}
				<CardContent>
					<div className="relative w-full aspect-square">
						<Image src="/placeholder-200x150.webp" alt={item.title || ""} className="rounded-lg mb-2 object-contain" fill />
					</div>
					<p className="font-semibold">當前出價：${item.currentBid}</p>
					<p className="text-sm text-gray-600">
						開始時間：{item.startTime.toLocaleString()}
					</p>
					<p className="text-sm text-gray-600">
						結束時間：{item.endTime.toLocaleString()}
					</p>
				</CardContent>
			</Card>
		</Link>
	);
}
