# DVD レンタル マイクロサービス

Go、gRPC、PostgreSQL で構築されたマイクロサービスベースの DVD レンタルシステム。

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

**[English](README.md)** | **[中文文档](README.zh-CN.md)**

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
- **2つの BFF サービス**: 顧客向けと管理者向けの REST API
- **JWT 認証**: デュアルトークン（Access + Refresh）と Redis 保存
- **全文検索**: PostgreSQL tsvector による映画検索
- **パーティションテーブル**: 月別決済パーティション
- **Docker 対応**: マルチステージ Alpine ビルド、docker-compose

## 技術スタック

| コンポーネント | 技術 |
|--------------|------|
| 言語 | Go 1.25+ |
| RPC | gRPC + Protocol Buffers v3 |
| HTTP | 標準ライブラリ `net/http` |
| データベース | PostgreSQL 17, sqlc |
| キャッシュ | Redis 7 |
| 認証 | JWT (golang-jwt/jwt/v5) |
| ビルド | Docker, docker-compose |

## クイックスタート

### 前提条件

- Go 1.25+
- Docker & Docker Compose
- buf（proto コード生成）
- sqlc（SQL コード生成）

### Docker で起動

```bash
# 全サービスを起動
docker compose -f deployments/docker-compose.yml up -d

# ログを確認
make infra-logs

# ステータスを確認
make infra-ps
```

### ローカル開発

```bash
# インフラのみ起動
make infra-up

# gRPC サービスを起動（各サービスを別ターミナルで）
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-store
DATABASE_URL="postgres://dvdrental:dvdrental@localhost:5432/dvdrental?sslmode=disable" make run-film
# ... 他のサービス

# BFF サービスを起動
JWT_SECRET="your-secret" REDIS_URL="redis://localhost:6379" make run-customer-bff
JWT_SECRET="your-secret" REDIS_URL="redis://localhost:6379" make run-admin-bff
```

## プロジェクト構成

```
dvd-rental/
├── cmd/                          # サービスエントリポイント
├── internal/                     # プライベートコード
│   ├── store/                    # 店舗サービス
│   ├── film/                     # 映画サービス
│   ├── customer/                 # 顧客サービス
│   ├── rental/                   # レンタルサービス
│   ├── payment/                  # 決済サービス
│   └── bff/                      # BFF サービス
├── pkg/                          # 共有パッケージ
├── proto/                        # Proto 定義
├── gen/                          # 生成されたコード
├── migrations/                   # データベースマイグレーション
└── deployments/                  # Docker 設定
```

## API エンドポイント

### 顧客 BFF（ポート 8080）

| エンドポイント | メソッド | 認証 | 説明 |
|--------------|--------|------|------|
| `/api/v1/auth/login` | POST | - | 顧客ログイン |
| `/api/v1/films` | GET | - | 映画一覧/検索 |
| `/api/v1/rentals` | GET/POST | JWT | レンタル管理 |
| `/api/v1/payments` | GET | JWT | 決済履歴 |
| `/api/v1/profile` | GET/PUT | JWT | プロフィール |

### 管理 BFF（ポート 8081）

| エンドポイント | メソッド | 認証 | 説明 |
|--------------|--------|------|------|
| `/api/v1/auth/login` | POST | - | スタッフログイン |
| `/api/v1/stores` | CRUD | JWT | 店舗管理 |
| `/api/v1/staff` | CRUD | JWT | スタッフ管理 |
| `/api/v1/customers` | CRUD | JWT | 顧客管理 |
| `/api/v1/films` | CRUD | JWT | 映画管理 |
| `/api/v1/inventory` | CRUD | JWT | 在庫管理 |
| `/api/v1/rentals` | CRUD | JWT | レンタル管理 |
| `/api/v1/payments` | CRUD | JWT | 決済管理 |

## 開発コマンド

```bash
make proto-gen     # protobuf コード生成
make sqlc-gen      # sqlc コード生成
make build-all     # 全サービスビルド
make test          # テスト実行
make fmt           # コードフォーマット
make lint          # リンター実行
```

## ライセンス

MIT License - 詳細は [LICENSE](LICENSE) を参照
