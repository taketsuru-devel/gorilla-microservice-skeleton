## 自分用skeleton

gorillaでslackbot作ってたらある程度共通の箇所が出てきたのでまとめ
slack周りのいい単体テストの方法が出てこなかったのでslack apiを使って試す

### memo

slackのevent subscriptionに登録したendpointはchallege処理が必要で関数化
bodyはjson形式
incoming webhookに登録したやつはchallenge処理不要
bodyはhttp form形式

### todo

動作確認
logはfilenameをもう少し縮めたい
