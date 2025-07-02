# Simple Kanban

Go (Gin) を使用したシンプルなKanbanボード管理システムのREST APIです。

## 機能

- **ユーザー認証**: JWT認証による登録・ログイン
- **ボード管理**: Kanbanボードの作成・編集・削除
- **カラム管理**: ボード内のカラム（To Do, In Progress, Done）
- **タスク管理**: タスクの作成・編集・削除・移動
- **ドラッグ&ドロップ**: タスクのカラム間移動と順序変更

## 技術スタック

### バックエンド
- **Go 1.21**: プログラミング言語
- **Gin**: Webフレームワーク
- **GORM**: ORMライブラリ
- **PostgreSQL**: データベース
- **JWT**: 認証システム
- **BCrypt**: パスワードハッシュ化

### インフラ
- **Docker**: コンテナ化
- **Docker Compose**: 開発環境

## プロジェクト構造

```
simple-kanban/
├── cmd/server/           # メインアプリケーション
├── internal/
│   ├── domain/          # エンティティ定義
│   ├── repository/      # データアクセス層
│   ├── service/         # ビジネスロジック層
│   └── handler/         # HTTPハンドラー層
├── pkg/
│   └── middleware/      # ミドルウェア（JWT認証など）
├── config/              # 設定管理
├── docker-compose.yml   # Docker構成
├── Dockerfile          # Dockerイメージ定義
└── README.md           # このファイル
```

## 起動方法

### 1. Docker Composeを使用（推奨）

```bash
# リポジトリをクローン
git clone <repository-url>
cd simple-kanban

# Docker Composeでサービスを起動
docker-compose up --build

# バックグラウンドで起動する場合
docker-compose up -d --build
```

### 2. ローカル環境での起動

#### 前提条件
- Go 1.21以上
- PostgreSQL 16
- Git

#### 手順

```bash
# 依存関係をインストール
go mod download

# PostgreSQLデータベースを作成
createdb simple_kanban

# 環境変数を設定（.env.exampleを参考に）
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=password
export DB_NAME=simple_kanban
export JWT_SECRET=your-secret-key

# アプリケーションを起動
# 1. PostgreSQL + APIサーバーをDockerで起動
docker-compose up -d

# 2. フロントエンドを起動
cd frontend && npm run dev

```

## API仕様

### ベースURL
```
http://localhost:8080/api/v1
```

### ヘルスチェック
```
GET /health
```

### 認証エンドポイント

#### ユーザー登録
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

#### ログイン
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "password123"
}
```

#### プロフィール取得
```http
GET /api/v1/auth/profile
Authorization: Bearer <JWT_TOKEN>
```

### ボードエンドポイント

#### ボード一覧取得
```http
GET /api/v1/boards
Authorization: Bearer <JWT_TOKEN>
```

#### ボード作成
```http
POST /api/v1/boards
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "name": "My Kanban Board"
}
```

#### ボード詳細取得（カラム・タスク付き）
```http
GET /api/v1/boards/:id/columns
Authorization: Bearer <JWT_TOKEN>
```

#### ボード更新
```http
PUT /api/v1/boards/:id
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "name": "Updated Board Name"
}
```

#### ボード削除
```http
DELETE /api/v1/boards/:id
Authorization: Bearer <JWT_TOKEN>
```

### タスクエンドポイント

#### タスク作成
```http
POST /api/v1/tasks
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "column_id": 1,
  "title": "新しいタスク",
  "description": "タスクの説明",
  "assignee_id": "user-uuid",
  "due_date": "2024-12-31T23:59:59Z"
}
```

#### タスク取得
```http
GET /api/v1/tasks/:id
Authorization: Bearer <JWT_TOKEN>
```

#### タスク更新
```http
PUT /api/v1/tasks/:id
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "title": "更新されたタスク",
  "description": "新しい説明"
}
```

#### タスク削除
```http
DELETE /api/v1/tasks/:id
Authorization: Bearer <JWT_TOKEN>
```

#### タスク移動（ドラッグ&ドロップ）
```http
PUT /api/v1/tasks/:id/move
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "new_column_id": 2,
  "new_order": 1
}
```

#### タスク順序変更
```http
PUT /api/v1/columns/:columnId/tasks/reorder
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "task_ids": [3, 1, 2]
}
```

## データベーススキーマ

### Users テーブル
- `id` (UUID, Primary Key)
- `email` (String, Unique)
- `password_hash` (String)
- `created_at` (Timestamp)
- `updated_at` (Timestamp)

### Boards テーブル
- `id` (Integer, Primary Key)
- `name` (String)
- `owner_id` (UUID, Foreign Key)
- `created_at` (Timestamp)
- `updated_at` (Timestamp)

### Columns テーブル
- `id` (Integer, Primary Key)
- `board_id` (Integer, Foreign Key)
- `title` (String)
- `order` (Integer)
- `created_at` (Timestamp)
- `updated_at` (Timestamp)

### Tasks テーブル
- `id` (Integer, Primary Key)
- `column_id` (Integer, Foreign Key)
- `title` (String)
- `description` (Text)
- `order` (Integer)
- `assignee_id` (UUID, Foreign Key, Optional)
- `due_date` (Timestamp, Optional)
- `created_at` (Timestamp)
- `updated_at` (Timestamp)

## 開発

### テスト実行
```bash
go test ./...
```

### ビルド
```bash
go build -o server cmd/server/main.go
```

### ホットリロード（開発時）
```bash
# air をインストール
go install github.com/cosmtrek/air@latest

# ホットリロードで起動
air
```

## 環境変数

| 変数名 | デフォルト値 | 説明 |
|--------|-------------|------|
| `PORT` | `8080` | サーバーポート |
| `GIN_MODE` | `debug` | Ginの実行モード |
| `DB_HOST` | `localhost` | データベースホスト |
| `DB_PORT` | `5432` | データベースポート |
| `DB_USER` | `postgres` | データベースユーザー |
| `DB_PASSWORD` | `password` | データベースパスワード |
| `DB_NAME` | `simple_kanban` | データベース名 |
| `JWT_SECRET` | `your-secret-key` | JWT署名キー |
| `JWT_EXPIRE_HOURS` | `24` | JWT有効期限（時間） |

## ライセンス

MIT License 