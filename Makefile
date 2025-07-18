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

# データベースのリセット
.PHONY: db-reset
db-reset:
	docker-compose down
	docker-compose up -d db
	sleep 5
	go run cmd/server/main.go

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

# 開発環境の起動（バックエンド + フロントエンド）
.PHONY: dev-full
dev-full:
	make dev &
	make frontend

# ヘルプ
.PHONY: help
help:
	@echo "利用可能なコマンド:"
	@echo "  make dev          - ホットリロードでサーバー起動"
	@echo "  make run          - 手動でサーバー起動"
	@echo "  make build        - プロダクションビルド"
	@echo "  make test         - テスト実行"
	@echo "  make test-coverage - テスト（カバレッジ付き）"
	@echo "  make lint         - リント実行"
	@echo "  make tidy         - 依存関係の整理"
	@echo "  make db-reset     - データベースのリセット"
	@echo "  make clean        - クリーンアップ"
	@echo "  make frontend     - フロントエンド起動"
	@echo "  make dev-full     - フルスタック開発環境起動" 