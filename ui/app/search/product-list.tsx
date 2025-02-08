'use client';

import { useEffect, useState, useCallback } from 'react';
import type { Defined, paths } from "@/app/openapi"
import {
    Pagination,
    PaginationContent,
    PaginationItem,
    PaginationLink,
    PaginationNext,
    PaginationPrevious,
} from "@/components/ui/pagination"
import createClient from "openapi-fetch";
import { BACKEND_API_BASE_URL } from '@/app/constants'
import { reParseJSON, dateReviver, serializeDeepObject } from '@/app/utils';
import { EnhancedGridContainer } from '@/app/components/enhanced-grid';
import { AuctionItem } from '@/app/components/auction-item';
import { useToast } from '@/hooks/use-toast';

type searchRequestType = Defined<paths["/auction/items"]["get"]["parameters"]["query"]>
type searchResultType = Defined<paths["/auction/items"]["get"]["responses"]["200"]["content"]["application/json"]["items"]>

type PageData = {
    items: searchResultType;
    nextCursor?: string;
}

export function ProductList({ searchRequest }: { searchRequest: searchRequestType }) {
    const [pages, setPages] = useState<PageData[]>([]);
    const [currentPage, setCurrentPage] = useState<PageData | undefined>(undefined);
    const [currentPageIndex, setCurrentPageIndex] = useState(0);
    const [maxPageIndex, setMaxPageIndex] = useState<number | undefined>(undefined);
    const [isLoading, setIsLoading] = useState(false);
    const { toast } = useToast()

    const withMinimumLoadingTime = useCallback(<TArgs extends unknown[], TReturn>(
        asyncFn: (...args: TArgs) => Promise<TReturn>,
        minimumTime = 500
    ) => {
        return async (...args: TArgs): Promise<TReturn> => {
            setIsLoading(true);
            const startTime = Date.now();
            try {
                return await asyncFn(...args);
            } catch (e) {
                throw e;
            } finally {
                const elapsedTime = Date.now() - startTime;
                const remainingTime = Math.max(0, minimumTime - elapsedTime);

                if (remainingTime > 0) {
                    await new Promise(resolve => setTimeout(resolve, remainingTime));
                    setIsLoading(false);
                }
            }
        };
    }, []);

    const fetchPage = useCallback(async (cursor?: string) => {
        const client = createClient<paths>({ baseUrl: BACKEND_API_BASE_URL });
        const { data, error, response } = await client.GET("/auction/items", {
            params: {
                query: {
                    ...searchRequest,
                    lastItemID: cursor,
                    size: 21,
                } as searchRequestType,
            },
            querySerializer: serializeDeepObject,
        });
        if (error && response.status !== 404) {
            toast({
                title: `查詢異常(${error.code})`,
                description: error.message,
                variant: "destructive",
            })
            console.error(`failed to fetch data with code ${error.code}, message: ${error.message}`);
            return undefined;
        }
        toast({
            title: `查詢成功`,
            description: `已取得${data?.count || 0}筆資料`,
        })
        if (!data) {
            return undefined;
        }
        const correctData = reParseJSON(data, dateReviver);
        if (!correctData || !correctData.items || correctData.items.length === 0) {
            return undefined;
        }
        return {
            items: correctData.items || [],
            nextCursor: correctData.items[correctData.items.length - 1].id,
        };
    }, [searchRequest, toast]);

    const handleNextPage = async () => {
        if (currentPageIndex < pages.length - 1) {
            const next = currentPageIndex + 1;
            setCurrentPageIndex(next);
            setCurrentPage(pages[next]);
            return;
        }
        if (currentPage && currentPage.nextCursor) {
            const nextPage = await fetchPage(currentPage.nextCursor);
            if (nextPage) {
                setPages([...pages, nextPage]);
                setCurrentPageIndex(currentPageIndex + 1);
                setCurrentPage(nextPage);
            } else {
                setMaxPageIndex(currentPageIndex + 1)
            }
        }
    };

    const visibleIndex = useCallback((total: number): Array<number> => {
        const start = Math.max(0, currentPageIndex - Math.floor(total / 2));
        const end = Math.min(pages.length, start + total);
        return Array.from({ length: end - start }, (_, i) => i + start);
    }, [currentPageIndex, pages]);

    useEffect(() => {
        const fetchFirstPage = async () => {
            const firstPage = await fetchPage();
            if (firstPage) {
                setPages([firstPage]);
                setCurrentPageIndex(0);
                setCurrentPage(firstPage);
            } else {
                setPages([]);
                setMaxPageIndex(0);
                setCurrentPage(undefined);
            }
        }
        withMinimumLoadingTime(fetchFirstPage)();
    }, [fetchPage, withMinimumLoadingTime]);

    return (
        <div className="flex-1">
            <EnhancedGridContainer minItemWidth={'300px'} maxItemWidth={'400px'} gap={'1rem'} className={`transition-opacity duration-300 ${isLoading ? 'opacity-50' : 'opacity-100'}`}>
                {currentPage?.items.map((item) => (
                    <AuctionItem key={item.id} item={item} />
                ))}
            </EnhancedGridContainer>

            <div className="flex justify-center mt-4">
                <Pagination>
                    <PaginationContent>
                        <PaginationItem>
                            <PaginationPrevious
                                onClick={withMinimumLoadingTime(async () => {
                                    const prev = Math.max(0, currentPageIndex - 1)
                                    console.log(`prev: ${currentPageIndex} to ${prev}`);
                                    setCurrentPageIndex(prev)
                                    setCurrentPage(pages[prev])
                                })}
                                aria-disabled={currentPageIndex < 1}
                                tabIndex={currentPageIndex < 1 ? -1 : undefined}
                                className={
                                    currentPageIndex < 1 ? "pointer-events-none opacity-50" : undefined
                                }
                            />
                        </PaginationItem>
                        {visibleIndex(5).map((index) => (
                            <PaginationItem key={index}>
                                <PaginationLink
                                    onClick={withMinimumLoadingTime(async () => {
                                        console.log('page', index);
                                        setCurrentPageIndex(index)
                                        setCurrentPage(pages[index])
                                    })}
                                    isActive={currentPageIndex === index}
                                >
                                    {index + 1}
                                </PaginationLink>
                            </PaginationItem>
                        ))}
                        <PaginationItem>
                            <PaginationNext
                                onClick={withMinimumLoadingTime(handleNextPage)}
                                aria-disabled={currentPageIndex >= pages.length - 1 && maxPageIndex !== undefined && currentPageIndex + 1 >= maxPageIndex}
                                tabIndex={currentPageIndex >= pages.length - 1 && maxPageIndex !== undefined && currentPageIndex + 1 >= maxPageIndex ? -1 : undefined}
                                className={
                                    currentPageIndex >= pages.length - 1 && maxPageIndex !== undefined && currentPageIndex + 1 >= maxPageIndex ? "pointer-events-none opacity-50" : undefined
                                }
                            />
                        </PaginationItem>
                    </PaginationContent>
                </Pagination>
            </div>
        </div>
    );
}

