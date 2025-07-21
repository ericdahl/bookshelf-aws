locals {
  recommendations_lambda_source_dir = "${path.module}/lambdas/recommendations"
  recommendations_go_files_for_hash = fileset(local.recommendations_lambda_source_dir, "**/*.go")
  recommendations_source_hash       = sha1(join("", [for f in local.recommendations_go_files_for_hash : filesha1("${local.recommendations_lambda_source_dir}/${f}")]))
}

resource "null_resource" "build_recommendations_lambda" {
  triggers = {
    source_hash = local.recommendations_source_hash
  }

  provisioner "local-exec" {
    command = "cd ${local.recommendations_lambda_source_dir} && make zip"
  }
}

data "aws_iam_policy_document" "recommendations_lambda_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "recommendations_lambda_exec_role" {
  name               = "recommendations-lambda-exec-role"
  assume_role_policy = data.aws_iam_policy_document.recommendations_lambda_assume_role_policy.json
}

data "aws_iam_policy_document" "recommendations_dynamodb_policy" {
  statement {
    actions   = ["dynamodb:Query"]
    resources = ["*"]
  }
}

data "aws_iam_policy_document" "recommendations_bedrock_policy" {
  statement {
    actions = [
      "bedrock:InvokeModel"
    ]
    resources = [
      "arn:aws:bedrock:*:*:foundation-model/amazon.titan-text-express-v1"
    ]
  }
}

resource "aws_iam_policy" "recommendations_dynamodb_policy" {
  name        = "RecommendationsDynamoDBPolicy"
  description = "Policy to allow querying the Books DynamoDB table"
  policy      = data.aws_iam_policy_document.recommendations_dynamodb_policy.json
}

resource "aws_iam_policy" "recommendations_bedrock_policy" {
  name        = "RecommendationsBedrockPolicy"
  description = "Policy to allow invoking Titan model in Bedrock"
  policy      = data.aws_iam_policy_document.recommendations_bedrock_policy.json
}

resource "aws_iam_role_policy_attachment" "recommendations_lambda_dynamodb_read" {
  role       = aws_iam_role.recommendations_lambda_exec_role.name
  policy_arn = aws_iam_policy.recommendations_dynamodb_policy.arn
}

resource "aws_iam_role_policy_attachment" "recommendations_lambda_bedrock_invoke" {
  role       = aws_iam_role.recommendations_lambda_exec_role.name
  policy_arn = aws_iam_policy.recommendations_bedrock_policy.arn
}

resource "aws_iam_role_policy_attachment" "recommendations_lambda_basic_execution" {
  role       = aws_iam_role.recommendations_lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_cloudwatch_log_group" "recommendations_lambda_log_group" {
  name              = "/aws/lambda/recommendations"
  retention_in_days = 7
}

resource "aws_lambda_function" "recommendations_lambda" {
  function_name = "recommendations"
  role          = aws_iam_role.recommendations_lambda_exec_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"
  timeout       = 30

  filename         = "${local.recommendations_lambda_source_dir}/dist/recommendations.zip"
  source_code_hash = local.recommendations_source_hash

  depends_on = [
    aws_iam_role_policy_attachment.recommendations_lambda_basic_execution,
    aws_iam_role_policy_attachment.recommendations_lambda_dynamodb_read,
    aws_iam_role_policy_attachment.recommendations_lambda_bedrock_invoke,
    null_resource.build_recommendations_lambda,
    aws_cloudwatch_log_group.recommendations_lambda_log_group,
  ]
}

resource "aws_apigatewayv2_integration" "recommendations_lambda_integration" {
  api_id           = aws_apigatewayv2_api.books_api.id
  integration_type = "AWS_PROXY"

  integration_uri        = aws_lambda_function.recommendations_lambda.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "recommendations_route" {
  api_id    = aws_apigatewayv2_api.books_api.id
  route_key = "GET /recommendations"
  target    = "integrations/${aws_apigatewayv2_integration.recommendations_lambda_integration.id}"

  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito_authorizer.id
}

resource "aws_lambda_permission" "recommendations_api_gateway_permission" {
  statement_id  = "AllowAPIGatewayInvokeRecommendations"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.recommendations_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.books_api.execution_arn}/*/*"
}