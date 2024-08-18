package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/cache/v9"
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

func main() {
	args := ParseArgs()
	sseServer, err := NewSSEServer(args.NatsURL)
	if err != nil {
		panic(err)
	}
	rdServer := redis.NewClient(&redis.Options{
		Addr: args.RedisURL,
	})
	if err := rdServer.Ping(context.Background()).Err(); err != nil {
		panic(fmt.Errorf("fail to connect to redis, err=%w", err))
	}
	httpServer := NewBidServer(rdServer, sseServer)
	httpServer.Run(args.ServerURL)
}

func NewBidServer(rdServer *redis.Client, sseServer *Event) *gin.Engine {
	p := bidProcessor{
		redServer: rdServer,
		redCache: cache.New(&cache.Options{
			Redis:      rdServer,
			LocalCache: cache.NewTinyLFU(1000, time.Minute),
		}),
		redSync: redsync.New(goredis.NewPool(rdServer)),
	}
	router := gin.Default()
	router.StaticFS("/public", http.Dir("./static/"))
	router.LoadHTMLGlob("templates/*")
	bidGroup := router.Group("/bid/item")
	{
		bidGroup.PUT("/:itemID/", func(c *gin.Context) {
			itemID := c.Param("itemID")
			if err := p.AddBidItem(c.Request.Context(), itemID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"detail": err.Error(),
				})
				return
			}
			c.Status(http.StatusOK)
		})
		bidGroup.DELETE("/:itemID/", func(c *gin.Context) {
			itemID := c.Param("itemID")
			if err := p.removeBidInfoFromRedis(c.Request.Context(), itemID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"detail": err.Error(),
				})
				return
			}
			if err := p.removeBidInfoFromDB(c.Request.Context(), itemID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"detail": err.Error(),
				})
				return
			}
			c.Status(http.StatusOK)
		})
		bidGroup.POST("/:itemID/", func(c *gin.Context) {
			itemID := c.Param("itemID")
			body := new(BidRequest)
			if err := c.ShouldBindBodyWithJSON(body); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"detail": err.Error(),
				})
				return
			}
			if err := p.Bid(c.Request.Context(), itemID, body); err != nil {
				out := gin.H{"detail": err.Error()}
				if errors.Is(err, ErrPriceTooLow) {
					c.JSON(http.StatusBadRequest, out)
				} else {
					c.JSON(http.StatusInternalServerError, out)
				}
				return
			}
			sseServer.Message <- BidNotification{
				ItemID: itemID,
				Info:   *body,
			}
			c.Status(http.StatusOK)
		})
		bidGroup.GET("/:itemID/info/", HeadersMiddleware(), sseServer.serveHTTP(), func(c *gin.Context) {
			itemID := c.Param("itemID")
			v, ok := c.Get("sseChannel")
			if !ok {
				return
			}
			clientChan, ok := v.(ClientChan)
			if !ok {
				return
			}
			var (
				info *BidRequest
				err  error
			)
			for {
				info, err = p.getBidInfoFromRedis(c.Request.Context(), itemID)
				if err == nil {
					break
				}
				c.SSEvent("error", gin.H{
					"detail": err.Error(),
				})
				time.Sleep(100 * time.Millisecond)
			}
			clientChan <- *info
		})
		bidGroup.GET("/:itemID/", func(c *gin.Context) {
			itemID := c.Param("itemID")
			_, err := p.getBidInfoFromRedis(c.Request.Context(), itemID)
			if errors.Is(err, redis.Nil) {
				c.Status(http.StatusNotFound)
				return
			} else if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"detail": err.Error(),
				})
				return
			}
			// c.Header("Cache-Control", "public, max-age=3600")
			c.HTML(http.StatusOK, "item.tmpl", gin.H{
				"itemID": itemID,
			})
		})
	}
	return router
}

var ErrPriceTooLow = errors.New("price too low")

type bidProcessor struct {
	// redis
	redServer *redis.Client
	redCache  *cache.Cache
	redSync   *redsync.Redsync
	// ignore RDBMS
}

func (p *bidProcessor) AddBidItem(ctx context.Context, itemID string) error {
	mutexKey := fmt.Sprintf("bid-%s-mutex", itemID)
	// 上分布式鎖
	m := p.redSync.NewMutex(mutexKey)
	if err := m.LockContext(ctx); err != nil {
		return fmt.Errorf("fail to lock bid info, itemID=%s", itemID)
	}
	// 寫入DB
	if err := p.addBidItemIntoDB(ctx, itemID); err != nil {
		return fmt.Errorf("fail to add item into database, err=%w", err)
	}
	// 寫入redis
	if err := p.addBidItemIntoRedis(ctx, itemID); err != nil {
		return fmt.Errorf("fail to initial bid info into redis, err=%w", err)
	}
	// 釋放分布式鎖
	if ok, err := m.UnlockContext(ctx); !ok {
		return fmt.Errorf("fail to unlock bid info")
	} else if err != nil {
		return fmt.Errorf("fail to unlock bid info, err=%w", err)
	}
	return nil
}

func (p *bidProcessor) Bid(ctx context.Context, itemID string, currentBid *BidRequest) error {
	mutexKey := fmt.Sprintf("bid-%s-mutex", itemID)
	// 短路徑:
	//  不使用分布式鎖的情況下，拿到的競拍資訊可能是舊的，可以快速的檢驗出價是否有高過之前的出價
	//  沒有超過的話，也不用去檢查當前的出價了
	prevBid, err := p.getBidInfoFromRedis(ctx, itemID)
	if err != nil {
		return fmt.Errorf("fail to get bid info without redlock, err=%w", err)
	}
	if prevBid.Bid >= currentBid.Bid {
		return ErrPriceTooLow
	}
	// 長路徑
	//  先獲取分布式鎖，確認拿到的競拍資料都是最新的
	//  以最新的價格來判斷出價是否超過之前的出價，並更新資料庫和快取
	m := p.redSync.NewMutex(mutexKey)
	if err := m.LockContext(ctx); err != nil {
		return fmt.Errorf("fail to lock bid info, itemID=%s", itemID)
	}
	info, err := p.getBidInfoFromRedis(ctx, itemID)
	if err != nil {
		return fmt.Errorf("fail to get bid info with redlock, err=%w", err)
	}
	if info.Bid >= currentBid.Bid {
		return ErrPriceTooLow
	}
	if err := p.setBidInfoIntoDB(ctx, itemID, currentBid); err != nil {
		return fmt.Errorf("fail to set bid info into database, err=%w", err)
	}
	if err := p.setBidInfoIntoRedis(ctx, itemID, currentBid); err != nil {
		return fmt.Errorf("fail to set bid info into redis, err=%w", err)
	}
	if ok, err := m.UnlockContext(ctx); !ok {
		return fmt.Errorf("fail to unlock bid info, itemID=%s", itemID)
	} else if err != nil {
		return fmt.Errorf("fail to unlock bid info, itemID=%s, err=%w", itemID, err)
	}
	return nil
}

func (p *bidProcessor) getBidInfoFromRedis(ctx context.Context, itemID string) (*BidRequest, error) {
	infoKey := fmt.Sprintf("bid-%s-info", itemID)
	info := new(BidRequest)
	if err := p.redServer.Get(ctx, infoKey).Scan(info); err != nil {
		return nil, fmt.Errorf("fail to get bid info from redis, err=%w", err)
	}
	return info, nil
}

func (p *bidProcessor) addBidItemIntoRedis(ctx context.Context, itemID string) error {
	infoKey := fmt.Sprintf("bid-%s-info", itemID)
	return p.redServer.Set(ctx, infoKey, new(BidRequest), 0).Err()
}

func (p *bidProcessor) addBidItemIntoDB(_ context.Context, _ string) error {
	return nil
}

func (p *bidProcessor) setBidInfoIntoRedis(ctx context.Context, itemID string, currentBid *BidRequest) error {
	infoKey := fmt.Sprintf("bid-%s-info", itemID)
	if err := p.redServer.Set(ctx, infoKey, currentBid, 0).Err(); err != nil {
		return fmt.Errorf("fail to set bid info into redis, err=%w", err)
	}
	return nil
}

func (p *bidProcessor) setBidInfoIntoDB(_ context.Context, _ string, _ *BidRequest) error {
	return nil
}

func (p *bidProcessor) removeBidInfoFromRedis(ctx context.Context, itemID string) error {
	infoKey := fmt.Sprintf("bid-%s-info", itemID)
	if err := p.redServer.Del(ctx, infoKey).Err(); err != nil {
		return fmt.Errorf("fail to delete bid info into redis, err=%w", err)
	}
	return nil
}

func (p *bidProcessor) removeBidInfoFromDB(_ context.Context, _ string) error {
	return nil
}
