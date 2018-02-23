[![CircleCI](https://circleci.com/gh/mmmpa/kick_my_mention.svg?style=svg)](https://circleci.com/gh/mmmpa/kick_my_mention)

# Kick my mention

現在時刻の 0 分から 1 時間前の 0 分までのメンションを採取し slack に通知する。

例えば 19:05 に起動すると 18:00 - 19:05 の範囲で取得する。

# Enviroment Variables

|key|value|
|:---|:---|
|LOCAL_RUN|ローカルテスト用。なければ AWS Lambda 用に `lambda.Start(execute)` として起動される|
|KICK_MY_MENTION_TOKEN|mention を取得する権限がある github token|
|KICK_MY_MENTION_SLACK_HOOK|Slack Incoming Webhooks の URL|