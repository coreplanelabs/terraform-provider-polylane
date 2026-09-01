data "aws_caller_identity" "current" {}

resource "polylane_aws_connection_request" "current" {
  workspace_id = "ws_00000000000000000000000000000000"
  account_id   = data.aws_caller_identity.current.account_id
  regions      = ["us-east-1", "us-west-2"]
}

data "aws_iam_policy_document" "polylane_trust" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = [polylane_aws_connection_request.current.principal_arn]
    }

    condition {
      test     = "StringEquals"
      variable = "sts:ExternalId"
      values   = [polylane_aws_connection_request.current.external_id]
    }
  }
}

resource "aws_iam_role" "polylane" {
  name               = "polylane-read"
  assume_role_policy = data.aws_iam_policy_document.polylane_trust.json
}

resource "aws_sns_topic_subscription" "polylane" {
  topic_arn = aws_sns_topic.cloudtrail.arn
  protocol  = "https"
  endpoint  = polylane_aws_connection_request.current.subscription_endpoint
}
