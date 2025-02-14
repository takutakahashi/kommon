# GitHub Webhook Integration Guide

このドキュメントでは、kommonでGitHub Webhookを使用する方法について説明します。

## 概要

kommonはGitHub Webhookを利用して、GitHubのイベントに応じて自動的にアクションを実行することができます。

## セットアップ手順

### 1. GitHub Webhookの設定

1. GitHubリポジトリの Settings > Webhooks に移動します
2. "Add webhook" をクリックします
3. 以下の情報を入力します：
   - Payload URL: `http://your-kommon-server:8080/webhook`
   - Content type: `application/json`
   - Secret: 任意の文字列（kommonの設定で使用します）

### 2. kommonの設定

kommonでGitHub Webhookを使用するには、以下のコマンドを実行します：

```bash
kommon gh webhook register \
  --endpoint http://your-kommon-server:8080/webhook \
  --secret your-webhook-secret
```

## イベントの設定

特定のイベントに対してアクションを設定するには：

```bash
kommon gh webhook set-event \
  --event pull_request \
  --action "your-command"
```

サポートされているイベント：
- pull_request
- push
- issue
- release

## 設定の確認

現在の設定を確認するには：

```bash
kommon gh webhook list
```

## トラブルシューティング

Webhookのデバッグ情報を確認するには：

```bash
kommon gh webhook debug
```

## セキュリティ注意事項

- Webhook Secretは必ず安全に保管してください
- エンドポイントはHTTPS使用を推奨します
- アクセス制限を適切に設定してください