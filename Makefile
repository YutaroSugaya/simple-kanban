# Simple Kanban Makefile

# 開発用コマンド
.PHONY: dev
dev:
	air

# 手動ビルド & 実行
.PHONY: run
run:
	go run cmd/server/main.go

# プロダクションビルド
.PHONY: build
build:
	go build -o bin/server cmd/server/main.go

# テスト実行
.PHONY: test
test:
	go test -v ./...

# テスト（カバレッジ付き）
.PHONY: test-coverage
test-coverage:
	go test -v -cover ./...

# リント実行
.PHONY: lint
lint:
	golangci-lint run

# 依存関係の整理
.PHONY: tidy
tidy:
	go mod tidy

# データベースの起動
.PHONY: db-start
db-start:
	docker-compose up -d db
	@echo "データベースの起動を待機中..."
	@until docker-compose exec db pg_isready -U postgres > /dev/null 2>&1; do \
		echo "PostgreSQLの起動を待機中..."; \
		sleep 2; \
	done
	@echo "PostgreSQLが起動しました！"

# データベースのリセット
.PHONY: db-reset
db-reset:
	docker-compose down
	docker-compose up -d db
	@echo "データベースの起動を待機中..."
	@until docker-compose exec db pg_isready -U postgres > /dev/null 2>&1; do \
		echo "PostgreSQLの起動を待機中..."; \
		sleep 2; \
	done
	@echo "PostgreSQLが起動しました！"

# データベースの停止
.PHONY: db-stop
db-stop:
	docker-compose down

# 全体のクリーンアップ
.PHONY: clean
clean:
	rm -rf bin/
	rm -rf tmp/
	docker-compose down

# フロントエンドの起動
.PHONY: frontend
frontend:
	cd frontend && npm run dev

# 開発環境の起動（データベース + バックエンド + フロントエンド）
.PHONY: dev-full
dev-full:
	@echo "🚀 Simple Kanban開発環境を起動しています..."
	@echo "1. データベースを起動中..."
	@make db-start
	@echo "2. バックエンドサーバーを起動中..."
	@make dev &
	@echo "3. フロントエンドサーバーを起動中..."
	@sleep 3
	@make frontend

# 開発環境の停止
.PHONY: dev-stop
dev-stop:
	@echo "開発環境を停止しています..."
	@pkill -f "air" || true
	@pkill -f "npm run dev" || true
	@make db-stop
	@echo "開発環境を停止しました"

# 依存関係のインストール
.PHONY: install
install:
	@echo "依存関係をインストールしています..."
	go mod tidy
	cd frontend && npm install
	@echo "依存関係のインストールが完了しました"

# ヘルプ
.PHONY: help
help:
	@echo "利用可能なコマンド:"
	@echo "  make install      - 依存関係のインストール"
	@echo "  make dev-full     - 完全な開発環境起動（DB+API+Frontend）"
	@echo "  make dev-stop     - 開発環境の停止"
	@echo "  make dev          - ホットリロードでサーバー起動"
	@echo "  make run          - 手動でサーバー起動"
	@echo "  make frontend     - フロントエンド起動"
	@echo "  make db-start     - データベース起動"
	@echo "  make db-stop      - データベース停止"
	@echo "  make db-reset     - データベースのリセット"
	@echo "  make build        - プロダクションビルド"
	@echo "  make test         - テスト実行"
	@echo "  make test-coverage - テスト（カバレッジ付き）"
	@echo "  make lint         - リント実行"
	@echo "  make tidy         - 依存関係の整理"
	@echo "  make clean        - クリーンアップ" 