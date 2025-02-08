'use server'

import {SearchBarProvider} from './searchbar-context'
import {SearchBar} from './searchbar'
import {ProductList} from "./product-list"
import {SearchBarActivator} from './hide-searchbar'
import type { Defined, paths } from "@/app/openapi"
import { EnhancedSearchParams } from '@/app/utils';

type searchRequestType = Defined<paths["/auction/items"]["get"]["parameters"]["query"]>

export default async function SearchPage({
    searchParams,
}: {
    searchParams: Promise<{ [key: string]: string | string[] | undefined }>
}) {
    const params = new EnhancedSearchParams(await searchParams);
    const searchRequest: searchRequestType = {
        title: params.getString('title'),
        startPrice: {
            from: params.getNumber('startPrice[from]'),
            to: params.getNumber('startPrice[to]'),
        },
        currentBid: {
            from: params.getNumber('currentBid[from]'),
            to: params.getNumber('currentBid[to]'),
        },
        startTime: {
            from: params.getDate('startTime[from]'),
            to: params.getDate('startTime[to]'),
        },
        endTime: {
            from: params.getDate('endTime[from]'),
            to: params.getDate('endTime[to]'),
        },
        sort: {
            key: params.getString('sort[key]') as "title" | "startPrice" | "currentBid" | "startTime" | "endTime" | undefined,
            order: params.getString('sort[order]') as "asc" | "desc" | undefined,
        },
        excludeEnded: params.getString('excludeEnded') === 'on' ? true : undefined,
    };
    return (
        <div className="container mx-auto px-4 py-8 gap-6">
            <SearchBarProvider>
                <div className="flex items-center justify-between mb-6">
                    <h1 className="text-3xl font-bold">商品搜尋</h1>
                    <div className="w-12 h-12">
                        <SearchBarActivator />
                    </div>
                </div>
                <div className="flex flex-col md:flex-row gap-6">
                    <SearchBar searchRequest={searchRequest} />
                    <ProductList searchRequest={searchRequest} />
                </div>
            </SearchBarProvider>
        </div>
    )
}
