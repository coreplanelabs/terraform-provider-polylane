resource "polylane_aws_connection" "current" {
  workspace_id           = polylane_aws_connection_request.current.workspace_id
  request_id             = polylane_aws_connection_request.current.id
  region                 = polylane_aws_connection_request.current.region
  role_arn               = aws_iam_role.polylane.arn
  bucket_name            = aws_s3_bucket.cloudtrail.id
  topic_arn              = aws_sns_topic.cloudtrail.arn
  topic_subscription_arn = aws_sns_topic_subscription.polylane.arn
  cloudtrail_name        = aws_cloudtrail.polylane.name
}
