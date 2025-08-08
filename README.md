# Simple Kanban

Go (Gin) + React を使用した高機能な Kanban ボード管理システムです。カレンダー表示、タイマー機能、GitHub 草風のアクティビティ表示など、プロダクティビティを向上させる機能を搭載しています。

## 🚀 機能

### 基本機能

- **ユーザー認証**: JWT 認証による登録・ログイン
- **ボード管理**: Kanban ボードの作成・編集・削除
- **カラム管理**: ボード内のカラム（To Do, In Progress, Done）
- **タスク管理**: タスクの作成・編集・削除・移動
- **ドラッグ&ドロップ**: タスクのカラム間移動と順序変更

### 新機能 ✨

- **📅 カレンダー表示**: 日単位・週間単位での切り替え表示

  - 10 分刻みの時間表示（カスタマイズ可能）
  - 縦型レイアウトで時間軸を表示
  - 平日・土日で異なる表示時間設定
  - タスクのドラッグ&ドロップでカレンダー配置

- **⏱️ タイマー機能**: ポモドーロテクニック対応

  - カウントダウンタイマー（設定時間から 0 まで）
  - 円形進捗バーによる視覚的表示
  - タスクと連動したタイマー管理
  - アクティブタイマーの状態表示

- **📊 タスク管理拡張**:

  - 目標時間の設定（分単位）
  - 実際の作業時間の自動計算
  - 完了状態の管理とビジュアル表示
  - スケジュール情報の保存

- **✅ 完了タスクの視覚表示**:

  - 完了時の斜線表示
  - 「Done」バッジの表示
  - 完了状態での透明度変更

- **🌱 GitHub 草風アクティビティ**:

  - 年間のタスク完了状況を視覚化
  - 完了数に応じた色の濃淡表示
  - 統計情報の表示（総完了数、アクティブ日数など）

- **⚙️ 設定画面**:
  - カレンダー表示時間のカスタマイズ
  - 平日・土日の個別設定
  - 時間スロット間隔の調整

## 🛠️ 技術スタック

### バックエンド

- **Go 1.21**: プログラミング言語
- **Gin**: Web フレームワーク
- **GORM**: ORM ライブラリ
- **PostgreSQL**: データベース
- **JWT**: 認証システム
- **BCrypt**: パスワードハッシュ化

### フロントエンド

- **React 18**: UI ライブラリ
- **TypeScript**: 型安全性
- **Vite**: ビルドツール
- **Tailwind CSS**: CSS フレームワーク
- **React Hook Form**: フォーム管理
- **React Router**: ルーティング
- **@hello-pangea/dnd**: ドラッグ&ドロップ
- **Axios**: HTTP 通信

### インフラ

- **Docker**: コンテナ化
- **Docker Compose**: 開発環境
- **Air**: ホットリロード（開発時）

## 📁 プロジェクト構造

```
simple-kanban/
├── cmd/server/                 # メインアプリケーション
├── internal/
│   ├── domain/                # エンティティ定義
│   │   ├── user.go
│   │   ├── board.go
│   │   ├── column.go
│   │   ├── task.go
│   │   ├── calendar_settings.go   # 新規追加
│   │   ├── timer_session.go       # 新規追加
│   │   └── calendar_event.go      # 新規追加
│   ├── repository/            # データアクセス層
│   ├── service/               # ビジネスロジック層
│   └── handler/               # HTTPハンドラー層
│       ├── auth_handler.go
│       ├── board_handler.go
│       ├── task_handler.go
│       ├── calendar_handler.go    # 新規追加
│       ├── timer_handler.go       # 新規追加
│       └── analytics_handler.go   # 新規追加
├── frontend/                  # React フロントエンド
│   ├── src/
│   │   ├── components/        # UIコンポーネント
│   │   │   ├── Calendar.tsx       # 新規追加
│   │   │   ├── Timer.tsx          # 新規追加
│   │   │   ├── TaskCard.tsx       # 機能拡張
│   │   │   └── ActivityHeatmap.tsx # 新規追加
│   │   ├── pages/             # ページコンポーネント
│   │   │   ├── Dashboard.tsx      # 機能拡張
│   │   │   ├── CalendarPage.tsx   # 新規追加
│   │   │   └── Settings.tsx       # 新規追加
│   │   ├── contexts/          # Reactコンテキスト
│   │   ├── types/            # TypeScript型定義
│   │   └── lib/              # ユーティリティ
├── pkg/
│   └── middleware/           # ミドルウェア（JWT認証など）
├── config/                   # 設定管理
├── docker-compose.yml        # Docker構成
├── Dockerfile               # Dockerイメージ定義
├── Makefile                 # 開発用コマンド
└── README.md               # このファイル
```

## 🚀 クイックスタート

### 1. 環境構築

```bash
# リポジトリをクローン
git clone <repository-url>
cd simple-kanban

# 依存関係をインストール
make install
```

### 2. 開発環境起動

```bash
# 完全な開発環境を起動（データベース + API + フロントエンド）
make dev-full
```

このコマンドは以下を実行します：

1. PostgreSQL データベースの起動と待機
2. Go API サーバーの起動（ホットリロード）
3. React 開発サーバーの起動

### 3. アクセス

- **フロントエンド**: http://localhost:5173
- **API**: http://localhost:8080
- **ヘルスチェック**: http://localhost:8080/health

### 4. 開発環境停止

```bash
make dev-stop
```

## 📝 Makefile コマンド

| コマンド        | 説明                                   |
| --------------- | -------------------------------------- |
| `make install`  | 依存関係のインストール                 |
| `make dev-full` | 完全な開発環境起動（DB+API+Frontend）  |
| `make dev-stop` | 開発環境の停止                         |
| `make dev`      | API サーバーのみ起動（ホットリロード） |
| `make frontend` | フロントエンドのみ起動                 |
| `make db-start` | データベース起動                       |
| `make db-stop`  | データベース停止                       |
| `make db-reset` | データベースリセット                   |
| `make build`    | プロダクションビルド                   |
| `make test`     | テスト実行                             |
| `make clean`    | クリーンアップ                         |

## 📖 使い方

### 1. アカウント作成とログイン

1. http://localhost:5173 にアクセス
2. 「新規登録」でアカウント作成
3. ダッシュボードが表示されます

### 2. ボードとタスクの作成

1. 「新しいボード」で Kanban ボードを作成
2. ボードを開いてタスクを追加
3. 目標時間や期限を設定

### 3. カレンダー機能の使用

1. ヘッダーの「カレンダー」をクリック
2. タスクをカレンダーにドラッグ&ドロップ
3. 「日」「週」ボタンで表示切り替え

### 4. タイマー機能の使用

1. タスクを選択してタイマーボタンをクリック
2. 時間を設定してタイマー開始
3. アクティブタイマーがヘッダーに表示

### 5. 設定のカスタマイズ

1. ヘッダーの「設定」をクリック
2. 平日・土日の表示時間を設定
3. 時間スロット間隔を調整

## 🔌 API 仕様

### ベース URL

```
http://localhost:8080/api/v1
```

### 新規 API エンドポイント

#### カレンダー関連

**カレンダー設定取得**

```http
GET /api/v1/calendar/settings
Authorization: Bearer <JWT_TOKEN>
```

**カレンダー設定更新**

```http
PUT /api/v1/calendar/settings
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "weekday_start_time": "09:00",
  "weekday_end_time": "18:00",
  "weekend_start_time": "10:00",
  "weekend_end_time": "16:00",
  "time_slot_duration": 10
}
```

**カレンダーイベント取得**

```http
GET /api/v1/calendar/events?start=2024-01-01T00:00:00Z&end=2024-01-07T23:59:59Z
Authorization: Bearer <JWT_TOKEN>
```

**タスクからカレンダーイベント作成**

```http
POST /api/v1/calendar/tasks/:taskId/events
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "start": "2024-01-15T10:00:00Z",
  "end": "2024-01-15T11:00:00Z"
}
```

#### タイマー関連

**タイマー開始**

```http
POST /api/v1/timer/start
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "task_id": 1,
  "duration": 1500
}
```

**タイマー停止**

```http
PUT /api/v1/timer/:id/stop
Authorization: Bearer <JWT_TOKEN>
```

**アクティブタイマー取得**

```http
GET /api/v1/timer/active
Authorization: Bearer <JWT_TOKEN>
```

#### 分析・統計

**タスク完了統計取得**

```http
GET /api/v1/analytics/task-completion?year=2024
Authorization: Bearer <JWT_TOKEN>
```

### 拡張されたタスク API

**タスク作成（拡張）**

```http
POST /api/v1/tasks
Authorization: Bearer <JWT_TOKEN>
Content-Type: application/json

{
  "column_id": 1,
  "title": "新しいタスク",
  "description": "タスクの説明",
  "estimated_time": 60,
  "due_date": "2024-12-31T23:59:59Z"
}
```

## 🗄️ データベーススキーマ

### 新規テーブル

#### CalendarSettings テーブル

- `id` (Integer, Primary Key)
- `user_id` (UUID, Foreign Key)
- `weekday_start_time` (String)
- `weekday_end_time` (String)
- `weekend_start_time` (String)
- `weekend_end_time` (String)
- `time_slot_duration` (Integer)
- `created_at` (Timestamp)
- `updated_at` (Timestamp)

#### TimerSessions テーブル

- `id` (Integer, Primary Key)
- `task_id` (Integer, Foreign Key)
- `user_id` (UUID, Foreign Key)
- `start_time` (Timestamp)
- `end_time` (Timestamp, Optional)
- `duration` (Integer)
- `is_active` (Boolean)
- `created_at` (Timestamp)
- `updated_at` (Timestamp)

#### CalendarEvents テーブル

- `id` (Integer, Primary Key)
- `user_id` (UUID, Foreign Key)
- `task_id` (Integer, Foreign Key, Optional)
- `title` (String)
- `start` (Timestamp)
- `end` (Timestamp)
- `color` (String)
- `is_task_based` (Boolean)
- `created_at` (Timestamp)
- `updated_at` (Timestamp)

### 拡張されたテーブル

#### Tasks テーブル（新規フィールド）

- `estimated_time` (Integer, Optional) - 目標時間（分）
- `actual_time` (Integer, Optional) - 実際の時間（分）
- `is_completed` (Boolean) - 完了状態
- `scheduled_start` (Timestamp, Optional) - スケジュール開始時刻
- `scheduled_end` (Timestamp, Optional) - スケジュール終了時刻
- `calendar_date` (Timestamp, Optional) - カレンダー配置日

## ⚙️ 環境変数

| 変数名              | デフォルト値      | 説明                         |
| ------------------- | ----------------- | ---------------------------- |
| `PORT`              | `8080`            | サーバーポート               |
| `GIN_MODE`          | `debug`           | Gin の実行モード             |
| `DB_HOST`           | `localhost`       | データベースホスト           |
| `DB_PORT`           | `5432`            | データベースポート           |
| `DB_USER`           | `postgres`        | データベースユーザー         |
| `DB_PASSWORD`       | `password`        | データベースパスワード       |
| `DB_NAME`           | `simple_kanban`   | データベース名               |
| `DB_SSL_MODE`       | `disable`         | SSL 接続モード               |
| `JWT_SECRET`        | `your-secret-key` | JWT 署名キー                 |
| `JWT_EXPIRE_HOURS`  | `24`              | JWT 有効期限（時間）         |
| `JWT_REFRESH_HOURS` | `168`             | JWT リフレッシュ期限（時間） |

## 🧪 開発・テスト

### テスト実行

```bash
make test
```

### ビルド

```bash
make build
```

### リント

```bash
make lint
```

## 🐳 Docker 環境

### 本番環境デプロイ

```bash
docker-compose up -d --build
```

### データベースのみ起動

```bash
make db-start
```

## 📸 スクリーンショット

### ダッシュボード

- GitHub 草風のアクティビティ表示
- ボード一覧表示

### カレンダー表示

- 日単位・週間単位の切り替え
- 10 分刻みの時間表示
- タスクのドラッグ&ドロップ

### タイマー機能

- 円形進捗バー
- タスクと連動したタイマー

### 設定画面

- 平日・土日の時間設定
- 時間スロットのカスタマイズ

## 🤝 コントリビューション

1. このリポジトリをフォーク
2. フィーチャーブランチを作成 (`git checkout -b feature/amazing-feature`)
3. 変更をコミット (`git commit -m 'Add some amazing feature'`)
4. ブランチをプッシュ (`git push origin feature/amazing-feature`)
5. プルリクエストを作成

## �� ライセンス

MIT License
