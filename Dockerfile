# ビルドステージ
FROM golang:1.23-alpine AS builder

# 作業ディレクトリを設定
WORKDIR /app

# go.modとgo.sumをコピー
COPY go.mod go.sum ./

# 依存関係をダウンロード
RUN go mod download

# ソースコードをコピー
COPY . .

# アプリケーションをビルド
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# 実行ステージ
FROM alpine:latest

# セキュリティと証明書のためのパッケージをインストール
RUN apk --no-cache add ca-certificates

# 作業ディレクトリを設定
WORKDIR /root/

# ビルドされたバイナリをコピー
COPY --from=builder /app/server .

# ポートを公開
EXPOSE 8080

# アプリケーションを実行
CMD ["./server"] 