# dynamodb-sort-uuid

ローカルのDynamoDBエミュレータ（`amazon/dynamodb-local`）に接続し、`user_id`ごとのItem一覧をJSONで返す簡易HTTPサーバー。DynamoDBのソートキー設計とUUIDの組み合わせを検証・学習する目的のプロジェクト。

本番でのテーブル作成はIaC（Terraform等）が担う前提であり、本リポジトリはテーブル作成に関与しない。認証情報・エンドポイントはコード内にローカル専用のダミー値を直書きしており、実AWSへの接続経路は持たない。

## テーブル設計

- テーブル名: `items`
- パーティションキー: `user_id`（String）
- ソートキー: `created_at`（String。値は`<ISO8601ナノ秒固定桁>#<job_id>`の連結文字列）
- 非キー属性: `job_id`（String、重複保持）

## セットアップ

```fish
make up             # DynamoDB Localをdocker composeで起動
make create-table   # テーブルを手動作成
make run             # サーバーを起動（:8080）
```

## Itemの投入（動作確認用）

`created_at`はソートキーの一部であり、`<ISO8601ナノ秒固定桁>#<job_id>`という連結文字列で保存する必要がある。`job_id`は非キー属性としても同じ値を重複保持する。

```fish
set USER_ID "test-user-1"
set CREATED_AT (date -u +%Y-%m-%dT%H:%M:%S).000000000Z
set JOB_ID (uuidgen)

aws dynamodb put-item \
  --endpoint-url http://localhost:8000 \
  --table-name items \
  --item "{\"user_id\":{\"S\":\"$USER_ID\"},\"created_at\":{\"S\":\"$CREATED_AT#$JOB_ID\"},\"job_id\":{\"S\":\"$JOB_ID\"}}" \
  --region us-east-1
```

保持窓（デフォルト24時間）より古いItemを投入して除外されることを確認したい場合は、`CREATED_AT`を過去の時刻にずらす。

```fish
set CREATED_AT (date -u -d '-30 hours' +%Y-%m-%dT%H:%M:%S).000000000Z
```

## API

```
GET /users/{user_id}/items
```

直近24時間分のItemをJSON配列で返す（保持窓はサーバー側の固定値、クライアントからの範囲指定は受け付けない）。

```sh
curl http://localhost:8080/users/<user_id>/items
```
