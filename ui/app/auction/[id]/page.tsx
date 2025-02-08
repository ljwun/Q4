'use server'

import { notFound } from 'next/navigation'
import { BACKEND_API_BASE_URL } from '@/app/constants'
import createClient from "openapi-fetch";
import type { paths } from "@/app/openapi/openapi"
import { AuctionItemInfo } from '@/app/auction/[id]/info'
import { reParseJSON, dateReviver } from '@/app/utils';

export default async function AuctionPage(props: { params: Promise<{ id: string }> }) {
    const params = await props.params;
    const client = createClient<paths>({ baseUrl: BACKEND_API_BASE_URL });
    const { data, error } = await client.GET("/auction/item/{itemID}", {
        params: {
            path: {
                itemID: params.id,
            },
        },
    });
    if (error || !data) {
        console.error('Failed to fetch auction item:', error);
        notFound();
    }
    return (
        <div className="container mx-auto px-4 py-8">
            <AuctionItemInfo id={params.id} info={reParseJSON(data, dateReviver)} />
        </div>
    )
}

