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

# deployment

Terraform でデプロイできる。以下の tf ファイルは各リソースへのアクセス権を付与した role を手作りすれば動作する (はず)。

```tf
variable "role" {}
variable "hook" {}
variable "token" {}

resource "aws_lambda_function" "kick_my_mention" {
  filename = "./main.zip"
  function_name = "kick_my_mention"
  role = "${var.role}"
  handler = "main"
  source_code_hash = "${base64sha256(file("./main.zip"))}"
  runtime = "go1.x"
  publish = false

  environment {
    variables = {
      KICK_MY_MENTION_SLACK_HOOK = "${var.hook}"
      KICK_MY_MENTION_TOKEN = "${var.token}"
    }
  }
}

resource "aws_cloudwatch_event_rule" "hourly_kick_mention" {
  name        = "hourly_kick_mention"
  description = "kick kick_my_mention hourly"
  schedule_expression = "cron(5 * * * ? *)"
  is_enabled = "true"
}

resource "aws_cloudwatch_event_target" "hourly_kick_mention_target" {
  target_id = "hourly_kick_mention_target"
  rule      = "${aws_cloudwatch_event_rule.hourly_kick_mention.name}"
  arn       = "${aws_lambda_function.kick_my_mention.arn}"
}

resource "aws_lambda_permission" "allow_hourly_kick_mention" {
  statement_id   = "lambda-allow_hourly_kick_mention"
  action         = "lambda:InvokeFunction"
  function_name  = "${aws_lambda_function.kick_my_mention.function_name}"
  principal      = "events.amazonaws.com"
  source_arn     = "${aws_cloudwatch_event_rule.hourly_kick_mention.arn}"
}

```
