import Link from 'next/link';
import { Button } from "@/components/ui/button";
import { ArrowRight } from 'lucide-react';
import { AuctionItem } from '@/app/components/auction-item';
import createClient from 'openapi-fetch';
import { BACKEND_API_BASE_URL } from '@/app/constants';
import type { Defined, paths } from '@/app/openapi';
import { reParseJSON, dateReviver, serializeDeepObject } from '@/app/utils';
import { EnhancedGridContainer } from './components/enhanced-grid';

type searchRequestType = Defined<paths["/auction/items"]["get"]["parameters"]["query"]>

async function getUpcomingAuctions() {
  const client = createClient<paths>({ baseUrl: BACKEND_API_BASE_URL });
  const { data, error, response } = await client.GET("/auction/items", {
    params: {
      query: { 
        startTime: { from: new Date()},
        sort: { key: "startTime", order: "asc" },      
        size: 20,
        excludeEnded: true,
      } as searchRequestType,
    },
    querySerializer: serializeDeepObject,
  });
  if (error && response.status !== 404) {
    console.error(`failed to fetch data with code ${error.code}, message: ${error.message}`);
    return undefined;
  }
  if (!data) {
    return undefined;
  }
  const correctData = reParseJSON(data, dateReviver);
  return correctData.items;
}

async function getEndingAuctions() {
  const client = createClient<paths>({ baseUrl: BACKEND_API_BASE_URL });
  const { data, error, response } = await client.GET("/auction/items", {
    params: {
      query: {
        sort: { key: "endTime", order: "asc" },
        size: 20,
        excludeEnded: true,
      } as searchRequestType,
    },
    querySerializer: serializeDeepObject,
  });
  if (error && response.status !== 404) {
    console.error(`failed to fetch data with code ${error.code}, message: ${error.message}`);
    return undefined;
  }
  if (!data) {
    return undefined;
  }
  const correctData = reParseJSON(data, dateReviver);
  return correctData.items;
}

export default async function HomePage() {
  const upcomingAuctions = await getUpcomingAuctions();
  const endingAuctions = await getEndingAuctions();

  return (
    <div className="container mx-auto px-4 py-8">
      <section className="text-center mb-16">
        <h1 className="text-5xl font-bold mb-4 text-foreground">探索多元拍賣世界</h1>
        <p className="text-xl text-muted-foreground mb-8">從電子產品到房地產，各種珍品等您競標</p>
        <Button asChild size="lg">
          <Link href="/search">
            開始探索 <ArrowRight className="ml-2 h-4 w-4" />
          </Link>
        </Button>
      </section>

      <section className="mb-16">
        <h2 className="text-3xl font-semibold mb-6 text-foreground">即將開始的拍賣</h2>
        <EnhancedGridContainer limitRows={1} minItemWidth={'300px'} maxItemWidth={'400px'} gap={'1.25rem'} className={`transition-opacity duration-300 opacity-100`}>
          {upcomingAuctions?.map((auction) => (
            <AuctionItem key={auction.id} item={auction} />
          ))}
        </EnhancedGridContainer>
      </section>

      <section className="mb-16">
        <h2 className="text-3xl font-semibold mb-6 text-foreground">即將結束的拍賣</h2>
        <EnhancedGridContainer limitRows={1} minItemWidth={'300px'} maxItemWidth={'400px'} gap={'1.25rem'} className={`transition-opacity duration-300 opacity-100`}>
          {endingAuctions?.map((auction) => (
            <AuctionItem key={auction.id} item={auction} />
          ))}
        </EnhancedGridContainer>
      </section>

      <section className="bg-muted rounded-lg p-8 text-center">
        <h2 className="text-3xl font-semibold mb-4 text-foreground">開始您的拍賣之旅</h2>
        <p className="text-muted-foreground mb-6">
          無論您是尋找獨特物品還是想出售珍藏，我們的平台都能滿足您的需求。
          立即加入我們，體驗刺激的競標過程！
        </p>
        <Button asChild size="lg">
          <Link href="/register">立即註冊</Link>
        </Button>
      </section>
    </div>
  );
}

