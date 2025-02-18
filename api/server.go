package api

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/redis/go-redis/v9"
	"github.com/samber/lo"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"q4/adapters/oidc"
	internalS3 "q4/adapters/s3"
	"q4/adapters/sse"
	"q4/api/openapi"
	"q4/models"
)

type ServerImpl struct {
	oidcProvider *oidc.ExtendedProvider
	sseManager   sse.ConnectionManager[openapi.BidEvent]
	s3Operator   *internalS3.S3Operator
	htmlChecker  *bluemonday.Policy
	db           *gorm.DB
}

func NewServer(config ServerConfig) (*ServerImpl, error) {
	const op = "NewServer"

	// 初始化OIDC提供者
	oidcProvider, err := oidc.NewExtendedProvider(config.OIDC.IssuerURL, config.OIDC.ClientID, config.OIDC.ClientSecret)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to initial OIDC provider, err=%w", op, err)
	}

	// 初始化S3客戶端
	s3Cfg, err := awsCfg.LoadDefaultConfig(
		context.Background(),
		awsCfg.WithBaseEndpoint(config.S3.Endpoint),
		awsCfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.S3.AccessKeyID, config.S3.SecretAccessKey, "")),
		awsCfg.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to load AWS config, err=%w", op, err)
	}
	s3Operator, err := internalS3.NewS3Operator(s3.NewFromConfig(s3Cfg), config.S3.Bucket, config.S3.PublicBaseURL)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to create S3 operator, err=%w", op, err)
	}

	// 初始化資料庫連線
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable&search_path=%s", config.DB.User, config.DB.Password, config.DB.Host, config.DB.Port, config.DB.Database, config.DB.Schema)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix: config.DB.Schema + ".",
		},
	})
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to connect to database, err=%w", op, err)
	}

	// 初始化Redis連線
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Addr,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	sseManager := sse.NewConnectionManager[openapi.BidEvent](redisClient, config.Redis.StreamKeys.SSE, slog.Default())
	sseManager.Start()

	return &ServerImpl{
		oidcProvider: oidcProvider,
		sseManager:   sseManager,
		s3Operator:   s3Operator,
		htmlChecker:  bluemonday.UGCPolicy(),
		db:           db,
	}, nil
}

func (impl *ServerImpl) Close() {
	impl.sseManager.Done()
}

// Add a new auction item
// (POST /auction/item)
func (impl *ServerImpl) PostAuctionItem(ctx context.Context, request openapi.PostAuctionItemRequestObject) (openapi.PostAuctionItemResponseObject, error) {
	const op = "PostAuctionItem"
	// 檢查拍賣物品的拍賣時間和結束時間是否合法
	if request.Body.StartTime.After(request.Body.EndTime) || request.Body.EndTime.Before(time.Now()) {
		return openapi.PostAuctionItem400JSONResponse{
			Message: lo.ToPtr("Invalid auction time"),
		}, nil
	}
	// 檢查使用者是否有權限新增拍賣物品
	if request.Params.AccessToken == nil {
		return openapi.PostAuctionItem401Response{}, nil
	}
	user, err := impl.oidcProvider.Introspect(*request.Params.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to introspect token, err=%w", op, err)
	}
	if !user.Active {
		return openapi.PostAuctionItem401Response{}, nil
	}
	// 處理拍賣描述
	if request.Body.Description != nil {
		request.Body.Description = lo.ToPtr(impl.htmlChecker.Sanitize(*request.Body.Description))
	}
	// 處理預設值
	if request.Body.Description == nil {
		request.Body.Description = lo.ToPtr("")
	}
	if request.Body.StartingPrice == nil {
		request.Body.StartingPrice = lo.ToPtr(int64(0))
	}
	if request.Body.StartTime == nil {
		request.Body.StartTime = lo.ToPtr(time.Now())
	}
	if request.Body.Carousels == nil {
		request.Body.Carousels = lo.ToPtr([]string{})
	}
	// 儲存拍賣物品
	dbUser := models.User{Username: user.Name}
	if result := impl.db.First(&dbUser); result.Error != nil {
		return nil, fmt.Errorf("[%s] Fail to find user, err=%w", op, result.Error)
	}
	auction := models.AuctionItem{
		UserID:        dbUser.ID,
		Title:         request.Body.Title,
		Description:   *request.Body.Description,
		StartingPrice: uint32(*request.Body.StartingPrice),
		CurrentBidID:  nil,
		StartTime:     *request.Body.StartTime,
		EndTime:       request.Body.EndTime,
		Carousels:     *request.Body.Carousels,
	}
	if result := impl.db.Debug().Create(&auction); result.Error != nil {
		return nil, fmt.Errorf("[%s] Fail to create auction item, err=%w", op, result.Error)
	}
	return openapi.PostAuctionItem201Response{
		Headers: openapi.PostAuctionItem201ResponseHeaders{
			Location: auction.ID.String(),
		},
	}, nil
}

// Get auction item details
// (GET /auction/item/{itemID})
func (impl *ServerImpl) GetAuctionItemItemID(ctx context.Context, request openapi.GetAuctionItemItemIDRequestObject) (openapi.GetAuctionItemItemIDResponseObject, error) {
	const op = "GetAuctionItemItemID"
	// 檢查拍賣物品是否存在
	auction := models.AuctionItem{ID: request.ItemID}
	if result := impl.db.Debug().
		Preload(
			"BidRecords",
			func(db *gorm.DB) *gorm.DB {
				return db.Order((clause.OrderByColumn{Column: clause.Column{Name: "created_at"}, Desc: true}))
			}).
		Preload("BidRecords.User").
		Preload("CurrentBid.User").
		First(&auction); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return openapi.GetAuctionItemItemID404Response{}, nil
		}
		return nil, fmt.Errorf("[%s] Fail to find auction item, err=%w", op, result.Error)
	}
	// 取得所有出價紀錄
	bidRecords := make([]openapi.BidEvent, len(auction.BidRecords))
	for i, bid := range auction.BidRecords {
		bidRecords[i] = openapi.BidEvent{
			Bid:  bid.Amount,
			User: bid.User.Username,
			Time: bid.CreatedAt,
		}
	}

	// 回傳拍賣物品資訊
	return openapi.GetAuctionItemItemID200JSONResponse{
		BidRecords:  bidRecords,
		Description: auction.Description,
		EndTime:     auction.EndTime,
		Title:       auction.Title,
		StartPrice:  int64(auction.StartingPrice),
		StartTime:   auction.StartTime,
		Carousels:   auction.Carousels,
	}, nil
}

// Place a bid on an auction item
// (POST /auction/item/{itemID}/bids)
func (impl *ServerImpl) PostAuctionItemItemIDBids(ctx context.Context, request openapi.PostAuctionItemItemIDBidsRequestObject) (openapi.PostAuctionItemItemIDBidsResponseObject, error) {
	const op = "PostAuctionItemItemIDBids"
	// 檢查拍賣物品是否存在
	auction := models.AuctionItem{ID: request.ItemID}
	if result := impl.db.First(&auction); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return openapi.PostAuctionItemItemIDBids400JSONResponse{}, nil
		}
		return nil, fmt.Errorf("[%s] Fail to find auction item, err=%w", op, result.Error)
	}
	// 檢查拍賣物品是否已經開始
	if time.Now().Before(auction.StartTime) {
		return openapi.PostAuctionItemItemIDBids403JSONResponse{}, nil
	}
	// 檢查拍賣物品是否已經結束
	if time.Now().After(auction.EndTime) {
		return openapi.PostAuctionItemItemIDBids410JSONResponse{}, nil
	}
	// 檢查使用者是否可以出價
	if request.Params.AccessToken == nil {
		return openapi.PostAuctionItemItemIDBids401Response{}, nil
	}
	user, err := impl.oidcProvider.Introspect(*request.Params.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to introspect token, err=%w", op, err)
	}
	if !user.Active {
		return openapi.PostAuctionItemItemIDBids401Response{}, nil
	}
	dbUser := models.User{Username: user.Name}
	if result := impl.db.First(&dbUser); result.Error != nil {
		return nil, fmt.Errorf("[%s] Fail to find user, err=%w", op, result.Error)
	}

	// 使用transaction確保出價的一致性
	errBidTooLow := errors.New("bid too low") // errBidTooLow用於標記出價太低的錯誤
	bidRecord := models.Bid{
		Amount:        uint32(request.Body.Bid),
		UserID:        dbUser.ID,
		AuctionItemID: auction.ID,
	}
	if err := impl.db.Transaction(func(tx *gorm.DB) error {
		// 1. 先寫入出價紀錄，減少商品資訊row lock的時間
		if result := tx.Create(&bidRecord); result.Error != nil {
			return fmt.Errorf("[%s] Fail to create bid, err=%w", op, result.Error)
		}
		// 2. 以單條SQL更新的方式來更新拍賣物品的最高出價，減少查詢一次所需的來回時間
		//    同時，避免讀取到SI(Snapshot Isolation)的副本，導致誤判(current_bid_id)，進一步造成Lost Update
		// todo: 增加最低出價間隔
		sub := tx.Model(&models.Bid{}).Select("1").Where("bids.id=current_bid_id").Where("bids.amount < ?", request.Body.Bid)
		result := tx.Model(&models.AuctionItem{}).Where("id = ?", auction.ID).Where(
			tx.Where("current_bid_id IS NULL").Or("EXISTS (?)", sub),
		).Update("current_bid_id", bidRecord.ID)
		if result.Error != nil {
			return fmt.Errorf("[%s] Fail to update auction item, err=%w", op, result.Error)
		}
		// 3. 透過UPDATE返回的更新數量，來判斷出價是否成功
		if result.RowsAffected == 0 {
			return errBidTooLow
		}
		return nil
	}); err != nil {
		if errors.Is(err, errBidTooLow) {
			return openapi.PostAuctionItemItemIDBids400JSONResponse{}, nil
		}
		return nil, fmt.Errorf("[%s] Fail to place bid, err=%w", op, err)
	}
	slog.Info("Higher bid occurs", slog.String("ID", bidRecord.ID.String()), slog.String("user", bidRecord.UserID.String()), slog.Int64("bid", int64(bidRecord.Amount)), slog.String("auctionID", bidRecord.AuctionItemID.String()))
	// 發送出價事件
	impl.sseManager.Publish(request.ItemID.String(), openapi.BidEvent{
		Bid:  request.Body.Bid,
		User: user.Nickname,
		Time: time.Now(),
	})
	return openapi.PostAuctionItemItemIDBids200Response{}, nil
}

// Track auction item events
// (GET /auction/item/{itemID}/events)
func (impl *ServerImpl) GetAuctionItemItemIDEvents(ctx context.Context, request openapi.GetAuctionItemItemIDEventsRequestObject) (openapi.GetAuctionItemItemIDEventsResponseObject, error) {
	const op = "GetAuctionItemItemIDEvents"
	// 檢查拍賣物品是否存在
	auction := models.AuctionItem{ID: request.ItemID}
	if result := impl.db.First(&auction); result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return openapi.GetAuctionItemItemIDEvents404Response{}, nil
		}
		return nil, fmt.Errorf("[%s] Fail to find auction item, err=%w", op, result.Error)
	}
	// 檢查拍賣物品是否已經開始拍賣(開始前5分鐘開放連線)
	if time.Now().Before(auction.StartTime.Add(-5 * time.Minute)) {
		return openapi.GetAuctionItemItemIDEvents403JSONResponse{
			Message: lo.ToPtr("Auction has not started"),
		}, nil
	}
	// 檢查拍賣物品是否已經結束拍賣
	if time.Now().After(auction.EndTime) {
		return openapi.GetAuctionItemItemIDEvents410JSONResponse{
			Message: lo.ToPtr("Auction has ended"),
		}, nil
	}
	// SSE請求合法，開始初始化串流
	c := ctx.(*gin.Context)
	w := c.Writer
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Transfer-Encoding", "chunked")
	ch, err := impl.sseManager.Subscribe(request.ItemID.String())
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to subscribe to item events, err=%w", op, err)
	}
LOOP:
	for {
		select {
		case <-w.CloseNotify():
			impl.sseManager.Unsubscribe(request.ItemID.String(), ch)
			break LOOP
		case event := <-ch:
			c.SSEvent("bid", event)
			w.Flush()
		// 30秒沒有事件就發送一個空行，確保瀏覽器和Cloudflare不會斷開連線
		case <-time.After(30 * time.Second):
			w.WriteString("\n\n")
			w.Flush()
		}
	}
	return openapi.GetAuctionItemItemIDEvents200Response{}, nil
}

// List auction items
// (GET /auction/items)
func (impl *ServerImpl) GetAuctionItems(ctx context.Context, request openapi.GetAuctionItemsRequestObject) (openapi.GetAuctionItemsResponseObject, error) {
	const op = "GetAuctionItems"
	now := time.Now()
	// 建立查詢
	query := impl.db.Debug().Joins("CurrentBid").Model(&models.AuctionItem{})
	//  - title
	if request.Params.Title != nil {
		query = query.Where("title LIKE ?", "%"+*request.Params.Title+"%")
	}
	//  - start_price
	if request.Params.StartPrice != nil {
		if request.Params.StartPrice.From != nil {
			query = query.Where("starting_price >= ?", *request.Params.StartPrice.From)
		}
		if request.Params.StartPrice.To != nil {
			query = query.Where("starting_price <= ?", *request.Params.StartPrice.To)
		}
	}
	//  - start_time
	if request.Params.StartTime != nil {
		if request.Params.StartTime.From != nil {
			query = query.Where("start_time >= ?", *request.Params.StartTime.From)
		}
		if request.Params.StartTime.To != nil {
			query = query.Where("start_time <= ?", *request.Params.StartTime.To)
		}
	}
	//  - end_time
	if request.Params.EndTime != nil {
		if request.Params.EndTime.From != nil {
			query = query.Where("end_time >= ?", *request.Params.EndTime.From)
		}
		if request.Params.EndTime.To != nil {
			query = query.Where("end_time <= ?", *request.Params.EndTime.To)
		}
	}
	//  - current_bid
	// 目前實際價格是記錄在另外一張表(bids)中，所以需要透過join來查詢
	// 且如果目前沒有人出價，則需要使用起標價格來進行篩選
	if request.Params.CurrentBid != nil {
		if request.Params.CurrentBid.From != nil {
			query = query.Where(`"CurrentBid".amount >= ? OR current_bid_id IS NULL AND starting_price >= ?`, *request.Params.CurrentBid.From, *request.Params.CurrentBid.From)
		}
		if request.Params.CurrentBid.To != nil {
			query = query.Where(`"CurrentBid".amount <= ? OR current_bid_id IS NULL AND starting_price <= ?`, *request.Params.CurrentBid.To, *request.Params.CurrentBid.To)
		}
	}
	//  - sort
	sortKey, desc := "title", false
	if request.Params.Sort != nil {
		if request.Params.Sort.Key != nil {
			switch *request.Params.Sort.Key {
			case openapi.Title:
				sortKey = "title"
			case openapi.StartTime:
				sortKey = "start_time"
			case openapi.EndTime:
				sortKey = "end_time"
			case openapi.CurrentBid:
				sortKey = "current_price"
			case openapi.StartPrice:
				sortKey = "starting_price"
			default:
				return openapi.GetAuctionItems400JSONResponse{
					Message: lo.ToPtr("Invalid sort key"),
				}, nil
			}
		}
		if request.Params.Sort.Order != nil {
			desc = *request.Params.Sort.Order == openapi.Desc
		}
	}
	query = query.Order(clause.OrderBy{Columns: []clause.OrderByColumn{
		{Column: clause.Column{Name: sortKey}, Desc: desc},
		{Column: clause.Column{Name: "id"}, Desc: false},
	}})
	//  - cursor
	if request.Params.LastItemID != nil {
		var cursor string
		if result := impl.db.Model(&models.AuctionItem{}).Select(sortKey).Where("id = ?", *request.Params.LastItemID).First(&cursor); result.Error != nil {
			if errors.Is(result.Error, gorm.ErrRecordNotFound) {
				return openapi.GetAuctionItems400JSONResponse{
					Message: lo.ToPtr("Last item not found"),
				}, nil
			}
			return nil, fmt.Errorf("[%s] Fail to find last item, err=%w", op, result.Error)
		}
		if desc {
			query = query.Where(sortKey+" < ?", cursor)
		} else {
			query = query.Where(sortKey+" > ?", cursor)
		}
		query = query.Or(sortKey+" = ? AND id > ?", cursor, *request.Params.LastItemID)
	}
	//  - size
	size := uint32(1)
	if request.Params.Size != nil {
		size = *request.Params.Size
	}
	query = query.Limit(int(size))
	//  - excludeEnded
	if request.Params.ExcludeEnded != nil && *request.Params.ExcludeEnded {
		query = query.Where("end_time > ?", now)
	}
	// todo: 嘗試從redis查詢，如果有就直接返回redis內儲存的查詢結果
	// 查詢拍賣物品
	var auctions []models.AuctionItem
	if result := query.Find(&auctions); result.Error != nil {
		return nil, fmt.Errorf("[%s] Fail to list auction items, err=%w", op, result.Error)
	}
	if len(auctions) == 0 {
		return openapi.GetAuctionItems404Response{}, nil
	}
	output := make([]struct {
		CurrentBid uint32    `json:"currentBid"`
		EndTime    time.Time `json:"endTime"`
		Id         uuid.UUID `json:"id"`
		IsEnded    bool      `json:"isEnded"`
		StartTime  time.Time `json:"startTime"`
		Title      string    `json:"title"`
	}, len(auctions))
	for i, auction := range auctions {
		if auction.CurrentBid != nil {
			output[i].CurrentBid = uint32(auction.CurrentBid.Amount)
		} else {
			output[i].CurrentBid = uint32(auction.StartingPrice)
		}
		output[i].Id = auction.ID
		output[i].Title = auction.Title
		output[i].EndTime = auction.EndTime
		output[i].StartTime = auction.StartTime
		output[i].IsEnded = now.After(auction.EndTime)
	}
	return openapi.GetAuctionItems200JSONResponse{
		Count: len(auctions),
		Items: output,
	}, nil
}

// Exchange authorization code
// (GET /auth/callback)
func (impl *ServerImpl) GetAuthCallback(ctx context.Context, request openapi.GetAuthCallbackRequestObject) (openapi.GetAuthCallbackResponseObject, error) {
	const op = "GetAuthCallback"
	// 驗證callback的參數和login時產生的參數是否相同
	reqestState, requestNonce := "", ""
	if request.Params.RequestState != nil {
		reqestState = *request.Params.RequestState
	}
	if request.Params.RequestNonce != nil {
		requestNonce = *request.Params.RequestNonce
	}
	verifier := impl.oidcProvider.NewExchangeVerifier(reqestState, requestNonce)
	// 向驗證伺服器交換token
	token, err := impl.oidcProvider.Exchange(ctx, verifier, request.Params.Code, request.Params.State, request.Params.RedirectUrl)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to exchange token, err=%w", op, err)
	}
	// 從id token中取得使用者資料
	var idTokenClaims oidc.UserInfo
	if err := json.Unmarshal(*token.IDTokenClaims, &idTokenClaims); err != nil {
		return nil, fmt.Errorf("[%s] Fail to retrieve email from id token, err=%w", op, err)
	}
	// 建立使用者資料(用於關聯使用者操作)
	var user models.User
	result := impl.db.FirstOrCreate(&user, &models.User{Username: idTokenClaims.Name})
	if result.Error != nil {
		return nil, fmt.Errorf("[%s] Fail to create user, err=%w", op, result.Error)
	}
	// 設定cookie並導向
	redirectUrl, err := url.Parse(request.Params.RedirectUrl)
	if err != nil {
		slog.Warn("[%s] Bad redirect url, err=%w", op, err)
	}
	return openapi.GetAuthCallback200Response{
		Headers: openapi.GetAuthCallback200ResponseHeaders{
			Location: redirectUrl.Query().Get("redirect_url"),
			SetCookieAccessTokenHttpOnlySecureMaxAge10800: token.OAuth2Token.AccessToken,
			SetCookieUsernameMaxAge10800:                  idTokenClaims.Nickname,
		},
	}, nil
}

// Obtain authentication url
// (GET /auth/login)
func (impl *ServerImpl) GetAuthLogin(ctx context.Context, request openapi.GetAuthLoginRequestObject) (openapi.GetAuthLoginResponseObject, error) {
	const op = "GetAuthLogin"
	state, err := generateID("st")
	if err != nil {
		return nil, fmt.Errorf("[%s] Unable to generate state, err=%w", op, err)
	}
	nonce, err := generateID("n")
	if err != nil {
		return nil, fmt.Errorf("[%s] Unable to generate nonce, err=%w", op, err)
	}
	return openapi.GetAuthLogin200Response{
		Headers: openapi.GetAuthLogin200ResponseHeaders{
			Location: impl.oidcProvider.AuthURL(state, nonce, request.Params.RedirectUrl, []string{"email", "openid", "profile"}),
			SetCookieRequestStateHttpOnlySecureMaxAge120: state,
			SetCookieRequestNonceHttpOnlySecureMaxAge120: nonce,
		},
	}, nil
}

// Revoke authentication token
// (GET /auth/logout)
func (impl *ServerImpl) GetAuthLogout(ctx context.Context, request openapi.GetAuthLogoutRequestObject) (openapi.GetAuthLogoutResponseObject, error) {
	// only clear the cookie without revoking the token
	return openapi.GetAuthLogout200Response{}, nil
}

// Upload an image
// (POST /image)
func (impl *ServerImpl) PostImage(ctx context.Context, request openapi.PostImageRequestObject) (openapi.PostImageResponseObject, error) {
	const op = "PostImage"
	// 檢查使用者是否可以上傳圖片
	if request.Params.AccessToken == nil {
		return openapi.PostImage401Response{}, nil
	}
	user, err := impl.oidcProvider.Introspect(*request.Params.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to introspect token, err=%w", op, err)
	}
	if !user.Active {
		return openapi.PostImage401Response{}, nil
	}
	// 限制圖片
	// 	1. 小於5MB
	// 	2. MIME類型為不包含腳本的圖片檔案
	body := internalS3.NewMaxSizeReader(request.Body, 5<<20)
	file, err := io.ReadAll(body)
	if errors.As(err, &internalS3.ErrReachLimitType) {
		return openapi.PostImage400JSONResponse{
			Message: lo.ToPtr(err.Error()),
		}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to read image, err=%w", op, err)
	}
	mimeType := http.DetectContentType(file)
	secure, ext := internalS3.CheckSecureImageAndGetExtension(mimeType)
	if !secure {
		return openapi.PostImage400JSONResponse{
			Message: lo.ToPtr(fmt.Sprintf("Invalid image type: %s", mimeType)),
		}, nil
	}
	// 透過S3 API儲存圖片
	url, err := impl.s3Operator.UploadFileToS3(ctx, uuid.New().String()+"."+ext, mimeType, file)
	if err != nil {
		return nil, fmt.Errorf("[%s] Fail to upload image, err=%w", op, err)
	}

	return openapi.PostImage201Response{
		Headers: openapi.PostImage201ResponseHeaders{
			Location: url,
		},
	}, nil
}

func generateID(prefix string) (string, error) {
	const op = "generateID"
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("[%s] Fail to generate unique id, err=%w", op, err)
	}
	return prefix + "_" + base64.URLEncoding.EncodeToString(bytes), nil
}
