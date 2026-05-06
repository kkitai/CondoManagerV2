# CondoManagerV2

マンション管理業務向けWebアプリケーション。クレーム管理・物件管理・ユーザー管理・AI支援機能を提供する。

## 技術スタック

- **言語**: Go 1.26
- **Webフレームワーク**: chi
- **DB**: PostgreSQL 16（pgx/v5）
- **マイグレーション**: goose
- **フロントエンド**: htmx + Tailwind CSS（CDN）
- **ロギング**: zap

## セットアップ

### 前提条件

- Go 1.26+（[goenv](https://github.com/go-nv/goenv) 推奨）
- Docker & Docker Compose

### 手順

```bash
# 1. 依存ライブラリのインストール
go mod download

# 2. 環境変数ファイルの作成
cp .env.example .env
# .env を編集して SESSION_SECRET などを設定

# 3. PostgreSQL の起動
make docker-up

# 4. マイグレーション実行
make migrate

# 5. サーバー起動
make run
```

サーバーは http://localhost:8080 で起動する。`GET /health` でヘルスチェック可能。

## アプリケーション実行方法

```bash
make run
```

または直接実行:

```bash
go run ./app/cmd/server/
```

## Unit Test 実行方法

```bash
make test
```

カバレッジレポートを生成する場合:

```bash
make test-coverage
# coverage.html が生成される
```

## E2E Test 実行方法

> E2EテストはPhase 1以降で順次追加予定。

## ディレクトリ構成

```
.
├── app/
│   ├── cmd/server/          # エントリーポイント (main.go)
│   ├── db/migrations/       # gooseマイグレーションファイル
│   ├── internal/
│   │   ├── config/          # 環境変数マッピング
│   │   ├── database/        # DB接続プール
│   │   ├── domain/          # ドメインモデル
│   │   ├── handler/         # HTTPハンドラー
│   │   ├── middleware/      # ミドルウェア
│   │   ├── repository/      # DBアクセス層
│   │   ├── service/         # ビジネスロジック
│   │   └── util/            # 共通ユーティリティ
│   ├── static/              # 静的ファイル
│   ├── templates/           # HTMLテンプレート（htmx）
│   └── uploads/             # アップロードファイル保存先
├── .env.example
├── docker-compose.yml
├── go.mod
└── Makefile
```

## Makefile コマンド一覧

| コマンド | 説明 |
|---|---|
| `make run` | サーバー起動 |
| `make build` | バイナリビルド（`bin/server`） |
| `make test` | ユニットテスト実行 |
| `make test-coverage` | テストカバレッジレポート生成 |
| `make migrate` | DBマイグレーション（up） |
| `make migrate-down` | DBマイグレーション（1件ロールバック） |
| `make migrate-status` | マイグレーション状態確認 |
| `make docker-up` | PostgreSQL起動 |
| `make docker-down` | PostgreSQL停止 |
| `make vet` | go vet 実行 |
