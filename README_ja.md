# DVD レンタル マイクロサービス

Go、gRPC、PostgreSQL で構築されたマイクロサービスベースの DVD レンタルシステム。

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**[English](README.md)** | **[中文文档](README_zh.md)**

## アーキテクチャ

```
┌─────────────────────────────────────────────────────────────────────┐
│                           クライアント                               │
│                  ┌──────────────┬──────────────┐                   │
│                  │  顧客アプリ  │   管理画面   │                   │
│                  └──────────────┴──────────────┘                   │
└─────────────────────────────────────────────────────────────────────┘
                         │                 │
                         ▼                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                        BFF レイヤー (REST)                          │
│            ┌────────────────┐    ┌────────────────┐                │
│            │ customer-bff   │    │   admin-bff    │                │
│            │    (:8080)     │    │    (:8081)     │                │
│            │ 17 エンドポイント│   │ 50 エンドポイント│               │
│            └────────────────┘    └────────────────┘                │
└─────────────────────────────────────────────────────────────────────┘
                         │                 │
                         ▼                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                     コアサービス層 (gRPC)                           │
│  ┌────────┐ ┌────────┐ ┌──────────┐ ┌────────┐ ┌─────────┐        │
│  │  店舗  │ │  映画  │ │   顧客   │ │ レンタル│ │  決済   │        │
│  │ :50051 │ │ :50052 │ │  :50053  │ │ :50054 │ │  :50055 │        │
│  └────────┘ └────────┘ └──────────┘ └────────┘ └─────────┘        │
└─────────────────────────────────────────────────────────────────────┘
                                │
                                ▼
┌─────────────────────────────────────────────────────────────────────┐
│                          データ層                                   │
│            ┌────────────────┐    ┌────────────────┐                │
│            │   PostgreSQL   │    │     Redis      │                │
│            │  (15 テーブル) │    │ (トークン保存) │                │
│            └────────────────┘    └────────────────┘                │
└─────────────────────────────────────────────────────────────────────┘
```

## 機能

- **5つの gRPC マイクロサービス**: 店舗、映画、顧客、レンタル、決済
- **2つの BFF サービス**: 顧客向け（17 エンドポイント）と管理者向け（50 エンドポイント）の REST API
- **JWT 認証**: デュアルトークン（Access 15分 + Refresh 7日）と Redis 保存
- **全文検索**: PostgreSQL tsvector による映画検索
- **パーティションテーブル**: 月別決済パーティション
- **Docker 対応**: マルチステージ Alpine ビルド、docker-compose

## 技術スタック

| コンポーネント | 技術 |
|--------------|------|
| 言語 | Go 1.25+ |
| RPC | gRPC + Protocol Buffers v3, buf |
| HTTP | 標準ライブラリ `net/http`（Go 1.22+ ServeMux）|
| データベース | PostgreSQL 17, sqlc |
| キャッシュ | Redis 7 |
| 認証 | JWT (golang-jwt/jwt/v5), bcrypt |
| ビルド | Docker, docker-compose |

## 前提条件

| ツール | 用途 | インストール |
|--------|------|------------|
| Go 1.25+ | 言語ランタイム | [go.dev/dl](https://go.dev/dl/) |
| Docker & Docker Compose | コンテナランタイム | [docs.docker.com](https://docs.docker.com/get-docker/) |
| buf | Protobuf コード生成 | `go install github.com/bufbuild/buf/cmd/buf@latest` |
| sqlc | SQL コード生成 | `go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest` |

## クイックスタート

### 方法 A: Docker で一括起動（最速）

全 7 サービス + PostgreSQL + Redis をコンテナでビルド・起動します。

```bash
# リポジトリをクローン
git clone https://github.com/enkaigaku/dvd-rental.git
cd dvd-rental

# 全サービスを起動
docker compose -f deployments/docker-compose.yml up -d

# 全サービスの稼働状況を確認
docker compose -f deployments/docker-compose.yml ps

# ログを確認
docker compose -f deployments/docker-compose.yml logs -f
```

起動後、以下のアドレスでアクセス可能です：
- 顧客 BFF: http://localhost:8080
- 管理 BFF: http://localhost:8081

### 方法 B: ローカル開発（日常開発に推奨）

インフラを Docker で、サービスをローカルで実行して高速にイテレーションします。

#### ステップ 1: インフラの起動

```bash
# PostgreSQL と Redis のみ起動
# データベースは migrations/ ディレクトリによりスキーマ＋サンプルデータが自動初期化されます
make infra-up
```

起動後：
- PostgreSQL: `localhost:5432`（ユーザー: `dvdrental`, パスワード: `dvdrental`, DB: `dvdrental`）
- Redis: `localhost:6379`

#### ステップ 2: Go 依存関係のインストール

```bash
go mod download
```

#### ステップ 3: gRPC サービスの起動

5つのターミナルを開き、各サービスを起動します：

```bash
# ターミナル 1 - 店舗サービス
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-store

# ターミナル 2 - 映画サービス
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-film

# ターミナル 3 - 顧客サービス
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-customer

# ターミナル 4 - レンタルサービス
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-rental

# ターミナル 5 - 決済サービス
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-payment
```

#### ステップ 4: BFF サービスの起動

さらに 2つのターミナルを開きます：

```bash
# ターミナル 6 - 顧客 BFF
JWT_SECRET="dev-secret-change-me" REDIS_URL="redis://localhost:6379" make run-customer-bff

# ターミナル 7 - 管理 BFF
JWT_SECRET="dev-secret-change-me" REDIS_URL="redis://localhost:6379" make run-admin-bff
```

#### ステップ 5: 動作確認

```bash
# 映画一覧を取得（認証不要）
curl http://localhost:8080/api/v1/films

# 顧客ログイン
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "mary.smith@sakilacustomer.org", "password": "password"}'
```

### サービスの停止

```bash
# Docker インフラを停止
make infra-down

# または全サービスを停止（方法 A の場合）
docker compose -f deployments/docker-compose.yml down
```

## プロジェクト構成

```
dvd-rental/
├── cmd/                          # サービスエントリポイント
│   ├── store-service/            #   店舗 gRPC サーバー
│   ├── film-service/             #   映画 gRPC サーバー
│   ├── customer-service/         #   顧客 gRPC サーバー
│   ├── rental-service/           #   レンタル gRPC サーバー
│   ├── payment-service/          #   決済 gRPC サーバー
│   ├── customer-bff/             #   顧客向け REST API
│   └── admin-bff/                #   管理者向け REST API
├── internal/                     # プライベートコード
│   ├── store/                    #   handler → service → repository
│   ├── film/                     #   handler → service → repository
│   ├── customer/                 #   handler → service → repository
│   ├── rental/                   #   handler → service → repository
│   ├── payment/                  #   handler → service → repository
│   └── bff/                      #   BFF レイヤー（HTTP → gRPC 変換）
│       ├── customer/             #     config + handler + router
│       └── admin/                #     config + handler + router
├── pkg/                          # 共有パッケージ
│   ├── auth/                     #   JWT、bcrypt、Refresh Token ストア
│   ├── middleware/               #   Auth、CORS、ロギング、Recovery、レスポンス
│   └── grpcutil/                 #   gRPC クライアント接続ヘルパー
├── proto/                        # Protocol Buffer 定義（.proto）
├── gen/                          # 生成されたコード
│   ├── proto/                    #   gRPC/protobuf 生成 Go コード
│   └── sqlc/                     #   sqlc 生成 Go コード
├── sql/                          # SQL クエリ定義（sqlc 用）
├── migrations/                   # データベーススキーマ + シードデータ
│   ├── 001_initial_schema.sql    #   コアテーブル定義
│   ├── 002_pagila_schema.sql     #   Pagila スキーマ拡張
│   ├── 003_pagila_data.sql       #   サンプルデータ
│   ├── 004_customer_auth.sql     #   顧客パスワードハッシュ
│   └── 005_staff_auth.sql        #   スタッフパスワードハッシュ
├── deployments/                  # Docker 設定
│   ├── docker-compose.yml
│   └── dockerfiles/              #   各サービスの Dockerfile
├── buf.yaml                      # Buf 設定
├── buf.gen.yaml                  # Buf コード生成設定
├── sqlc.yaml                     # sqlc 設定
├── Makefile                      # 開発コマンド
└── go.mod
```

## API エンドポイント

### 顧客 BFF（ポート 8080）

| メソッド | パス | 認証 | 説明 |
|---------|------|------|------|
| POST | `/api/v1/auth/login` | - | 顧客ログイン |
| POST | `/api/v1/auth/refresh` | - | トークンリフレッシュ |
| POST | `/api/v1/auth/logout` | - | ログアウト |
| GET | `/api/v1/films` | - | 映画一覧 |
| GET | `/api/v1/films/search` | - | 映画検索 |
| GET | `/api/v1/films/category/{id}` | - | カテゴリ別映画 |
| GET | `/api/v1/films/actor/{id}` | - | 俳優別映画 |
| GET | `/api/v1/films/{id}` | - | 映画詳細 |
| GET | `/api/v1/categories` | - | カテゴリ一覧 |
| GET | `/api/v1/actors` | - | 俳優一覧 |
| GET | `/api/v1/rentals` | JWT | マイレンタル |
| GET | `/api/v1/rentals/{id}` | JWT | レンタル詳細 |
| POST | `/api/v1/rentals` | JWT | レンタル作成 |
| POST | `/api/v1/rentals/{id}/return` | JWT | 返却 |
| GET | `/api/v1/payments` | JWT | 決済履歴 |
| GET | `/api/v1/profile` | JWT | マイプロフィール |
| PUT | `/api/v1/profile` | JWT | プロフィール更新 |

### 管理 BFF（ポート 8081）

| メソッド | パス | 認証 | 説明 |
|---------|------|------|------|
| POST | `/api/v1/auth/login` | - | スタッフログイン |
| POST | `/api/v1/auth/refresh` | - | トークンリフレッシュ |
| POST | `/api/v1/auth/logout` | - | ログアウト |
| | `/api/v1/stores/**` | JWT | 店舗管理（CRUD）|
| | `/api/v1/staff/**` | JWT | スタッフ管理（CRUD）|
| | `/api/v1/customers/**` | JWT | 顧客管理（CRUD）|
| | `/api/v1/films/**` | JWT | 映画管理（CRUD）|
| | `/api/v1/actors/**` | JWT | 俳優管理（CRUD）|
| | `/api/v1/categories/**` | JWT | カテゴリ管理（CRUD）|
| | `/api/v1/languages/**` | JWT | 言語管理（CRUD）|
| | `/api/v1/inventory/**` | JWT | 在庫管理（CRUD）|
| | `/api/v1/rentals/**` | JWT | レンタル管理（CRUD）|
| | `/api/v1/payments/**` | JWT | 決済管理（CRUD）|

## 環境変数

### gRPC サービス（store、film、customer、rental、payment）

| 変数 | 必須 | デフォルト | 説明 |
|------|------|----------|------|
| `DATABASE_URL` | はい | - | PostgreSQL 接続文字列 |
| `GRPC_PORT` | いいえ | サービス固有 | gRPC リッスンポート（50051-50055）|
| `LOG_LEVEL` | いいえ | `info` | ログレベル（debug, info, warn, error）|

### BFF サービス（customer-bff、admin-bff）

| 変数 | 必須 | デフォルト | 説明 |
|------|------|----------|------|
| `HTTP_PORT` | いいえ | `8080` / `8081` | HTTP リッスンポート |
| `JWT_SECRET` | はい | - | JWT 署名シークレット |
| `JWT_ACCESS_DURATION` | いいえ | `15m` | Access Token 有効期間 |
| `JWT_REFRESH_DURATION` | いいえ | `168h`（7日）| Refresh Token 有効期間 |
| `REDIS_URL` | はい | - | Redis 接続 URL（Refresh Token 保存用）|
| `GRPC_STORE_ADDR` | いいえ | `localhost:50051` | 店舗サービスアドレス（admin-bff のみ）|
| `GRPC_FILM_ADDR` | いいえ | `localhost:50052` | 映画サービスアドレス |
| `GRPC_CUSTOMER_ADDR` | いいえ | `localhost:50053` | 顧客サービスアドレス |
| `GRPC_RENTAL_ADDR` | いいえ | `localhost:50054` | レンタルサービスアドレス |
| `GRPC_PAYMENT_ADDR` | いいえ | `localhost:50055` | 決済サービスアドレス |
| `LOG_LEVEL` | いいえ | `info` | ログレベル |

## 開発コマンド

```bash
# コード生成
make proto-gen          # protobuf Go コード生成（buf が必要）
make sqlc-gen           # sqlc Go コード生成（sqlc が必要）
make generate           # proto-gen と sqlc-gen を両方実行

# ビルド
make build-all          # 全 7 サービスをビルド

# テスト & リント
make test               # テスト実行（race 検出器有効）
make lint               # golangci-lint 実行
make fmt                # コードフォーマット（gofmt + goimports）

# インフラ（Docker）
make infra-up           # PostgreSQL + Redis を起動
make infra-down         # インフラを停止
make infra-logs         # インフラログを表示
make infra-ps           # インフラ状況を表示

# サービス起動（ローカル）
make run-store          # 店舗サービスを起動
make run-film           # 映画サービスを起動
make run-customer       # 顧客サービスを起動
make run-rental         # レンタルサービスを起動
make run-payment        # 決済サービスを起動
make run-customer-bff   # 顧客 BFF を起動
make run-admin-bff      # 管理 BFF を起動
```

## ライセンス

MIT License - 詳細は [LICENSE](LICENSE) を参照
