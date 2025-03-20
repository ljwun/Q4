# Q4

[![Tag](https://img.shields.io/github/v/tag/ljwun/q4?label=latest)](https://github.com/GreptimeTeam/greptimedb/releases/latest) ![GoVersion](https://img.shields.io/github/go-mod/go-version/ljwun/q4) ![NextJSVersion](https://img.shields.io/github/package-json/dependency-version/ljwun/q4/next?filename=ui/package.json) ![License](https://img.shields.io/github/license/ljwun/q4)

專案的預設目標是分佈式拍賣系統，目的是展示**資料的兢爭處理**和**分佈式架構設計**。

---

## Feature

- 使用 Server-Sent Events (SSE) 實現即時出價通知
- 使用 Redis 實現出價動作的快速回應
- 使用 Redis Stream 實現出價紀錄同步到資料庫
- 使用分布式鎖確保 Redis Stream 內的資料同一時間只有一個 Server 會將出價紀錄進行同步
- 實現自動續期的分布式鎖，保證處理同步作業的 Server 在下線時，快速釋放分布式鎖
- API 提供 Cursor Pagination 的搜尋方式，提升搜尋拍賣商品的效能
- 使用 S3 API 將使用者上傳的圖片儲存到 Cloudflare R2
- 提供 OIDC 的登入方法
- 提供 Dockerfile 和 Helm Chart 方便部署

## Change Log

- v0.6.1 🐛修正未登入的提示訊息造成 UI 錯誤的問題
- v0.6.0 🆕支援多平台的 SSO 登入，以及基本的用戶管理功能
- v0.5.0 🆕紀錄並限制使用者上傳圖片的頻率
- v0.4.0 🔄改由 Server 產生 JWT，提升 API 處理時驗證使用者身分的速度
- v0.3.3 🐛修正日誌寫入時調用到可能會是空值的當前出價欄位
- v0.3.2 🐛修正新增拍賣時的圖片網址寫入資料庫造成錯誤的問題
- v0.3.1 🐛修正 Helm Chart 中，不同 Helm Release 但會共用 Pod Label 的問題
- v0.3.0 🔄實現異步出價，由 Redis 比價後快速回應使用者，系統再以異步方式同步到資料庫
- v0.2.0 🆕使用 Redis Stream 來同步跨 Server 的出價處理
- v0.1.1 🐛修正 Helm Chart 中 Service 使用 NodePort 多暴露了埠的問題，改為 ClusterIP
- v0.1.0 🆕初版

## Demo Site

- 網站網址：[https://bid-demo-1.yyuk-dev.link](https://bid-demo-1.yyuk-dev.link)
- 系統狀態：![DemoServerStatus](https://img.shields.io/website?label=&url=https://bid-demo-1.yyuk-dev.link)
- 使用限制:
  - ⚠️每小時最多上傳 2 張圖片
  - ⚠️上傳的圖片隨時可能被刪除
  - ⚠️Internal SSO Provider 不提供註冊，但能使用 Google 登入

## Prepare

### Database Migration

1. 安裝 Atlas

   請參考[ariga/atlas](https://github.com/ariga/atlas/releases)。

2. 準備 `models/.env`

    ```bash
    npm run setup:models
    ```

    或者

    ```bash
    cp -u ./models/.env.sample ./models/.env
    ```

3. 修改 `models/.env` 內的 `TARGET_DB_DSN`

    :bulb: `DEV_DB_DSN` 變數只有在生成新的 migration 時會需要。

4. 執行遷移

    ```bash
    npm run migrate:prod
    ```

    或者

    ```bash
    export $(grep -v '^#' models/.env | xargs)
    atlas migrate apply -c file://models/atlas.hcl --env gorm
    ```

### Create Redis Stream and Consumer Group

根據需求設定 `<stream>` 和 `<group>`，需要反映到 `.env` 的 `Q4_REDIS_STREAM_KEY_FOR_BID` 和 `Q4_REDIS_CONSUMER_GROUP` 上。

```bash
redis-cli XGROUP CREATE <stream> <group> 0 MKSTREAM
```

也可以使用預設的名稱來建立:

```bash
redis-cli XGROUP CREATE q4-shared-bid-stream q4-bid-group 0 MKSTREAM
```

## Quick Start

### Helm Chart

1. 下載指定版本的 Helm Chart 預設值

    ```bash
    helm show values oci://ghcr.io/ljwun/helm/q4 --version <version> > q4-helm-chart-value.yaml
    ```

2. 修改 `q4-helm-chart-value.yaml`

3. 使用修改的設定檔來安裝指定版本的 Q4

    ```bash
    helm install my-release oci://ghcr.io/ljwun/helm/q4 --version <version> -f q4-helm-chart-value.yaml
    ```

### Docker

1. 下載專案

    ```bash
    git clone --depth 1 --branch <version> https://github.com/ljwun/Q4.git
    cd ./Q4
    ```

2. 準備設定檔

    :bulb: 以下指令會產生三個預設設定檔 `./.env`  `./ui/.env`  `./models/.env`

    ```bash
    npm run setup:all
    ```

3. 修改設定檔 `./.env` 和 `./ui/.env`

4. 準備 docker-compose

    :bulb: 需要配合實際狀況修改 `<service port>` 和 `<service ip>` 。

    ```yaml
    services:
      ui:
        image: ghcr.io/ljwun/q4-ui:latest
        build:
          context: ../../
          dockerfile: .ci/ui.Dockerfile
        ports:
          - "<service port>:3000"
        env_file:
          - ../../ui/.env
        environment:
          - HOSTNAME=0.0.0.0
          - PORT=3000
          - Q4_FRONTEND_BASE_URL=http://<service ip>:<service port>/
          - Q4_BACKEND_BASE_URL=http://api:8080/
        depends_on:
          - api
        networks:
          - q4-network
      api:
        image: ghcr.io/ljwun/q4-api:latest
        build:
          context: ../../
          dockerfile: .ci/api.Dockerfile
        env_file:
          - ../../.env
        environment:
          - Q4_SERVER_URL=0.0.0.0:8080
        networks:
          - q4-network
    networks:
      q4-network:
        driver: bridge
    ```

5. 啟動服務

    ```bash
    docker compose up -d
    ```

### Local

1. 下載專案

    ```bash
    git clone --depth 1 --branch <version> https://github.com/ljwun/Q4.git
    cd ./Q4
    ```

2. 安裝依賴

    ```bash
    npm run install:all
    ```

3. 準備設定檔

    :bulb: 以下指令會產生三個預設設定檔 `./.env`  `./ui/.env`  `./models/.env`

    ```bash
    npm run setup:all
    ```

4. 修改設定檔 `./.env` 和 `./ui/.env`

5. 啟動 api server 和 ui server

    ```bash
    npm run start:all
    ```

## Mechanism

### Bidding (待補)

### Bid Synchronous (待補)

## License

本專案採用 [Apache License 2.0](LICENSE) 授權。任何人都可以自由使用、修改和分發本程式碼，但必須保留原始版權聲明並標明修改。

詳細條款請參閱 [LICENSE](LICENSE) 檔案。
