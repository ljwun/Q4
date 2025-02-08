# 第一階段: 建立環境進行編譯
FROM golang:1.23.4 AS builder

# 設置工作目錄
WORKDIR /app

# 將 go.mod 和 go.sum 複製到容器中並下載依賴
COPY go.mod go.sum ./
RUN go mod download

# 複製其他程式碼並進行編譯
COPY . ./
RUN go build -o main .

# 第二階段: 建立運行環境
FROM gcr.io/distroless/base-debian12 AS runtime

# # 將二進制文件從 builder 複製到 runtime 階段
COPY --from=builder /app/main /main

# 設定執行時的默認指令
ENTRYPOINT ["/main"]
