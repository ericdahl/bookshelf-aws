locals {
  get_book_lambda_source_dir = "${path.module}/lambdas/get-book"
  get_book_go_files_for_hash = fileset(local.get_book_lambda_source_dir, "**/*.go")
  get_book_source_hash       = sha1(join("", [for f in local.get_book_go_files_for_hash : filesha1("${local.get_book_lambda_source_dir}/${f}")]))
}

resource "null_resource" "build_get_book_lambda" {
  triggers = {
    source_hash = local.get_book_source_hash
  }

  provisioner "local-exec" {
    command = "cd ${local.get_book_lambda_source_dir} && make zip"
  }
}

data "aws_iam_policy_document" "get_book_lambda_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "get_book_lambda_exec_role" {
  name               = "get-book-lambda-exec-role"
  assume_role_policy = data.aws_iam_policy_document.get_book_lambda_assume_role_policy.json
}

data "aws_iam_policy_document" "get_book_dynamodb_policy" {
  statement {
    actions   = ["dynamodb:GetItem"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "get_book_dynamodb_policy" {
  name        = "GetBookDynamoDBPolicy"
  description = "Policy to allow getting an item from the Books DynamoDB table"
  policy      = data.aws_iam_policy_document.get_book_dynamodb_policy.json
}

resource "aws_iam_role_policy_attachment" "get_book_lambda_dynamodb_read" {
  role       = aws_iam_role.get_book_lambda_exec_role.name
  policy_arn = aws_iam_policy.get_book_dynamodb_policy.arn
}

resource "aws_iam_role_policy_attachment" "get_book_lambda_basic_execution" {
  role       = aws_iam_role.get_book_lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_cloudwatch_log_group" "get_book_lambda_log_group" {
  name              = "/aws/lambda/get-book"
  retention_in_days = 7
}

resource "aws_lambda_function" "get_book_lambda" {
  function_name = "get-book"
  role          = aws_iam_role.get_book_lambda_exec_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"

  filename         = "${local.get_book_lambda_source_dir}/dist/get-book.zip"
  source_code_hash = local.get_book_source_hash

  depends_on = [
    aws_iam_role_policy_attachment.get_book_lambda_basic_execution,
    aws_iam_role_policy_attachment.get_book_lambda_dynamodb_read,
    null_resource.build_get_book_lambda,
    aws_cloudwatch_log_group.get_book_lambda_log_group,
  ]
}

resource "aws_apigatewayv2_integration" "get_book_lambda_integration" {
  api_id           = aws_apigatewayv2_api.books_api.id
  integration_type = "AWS_PROXY"

  integration_uri        = aws_lambda_function.get_book_lambda.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "get_book_route" {
  api_id    = aws_apigatewayv2_api.books_api.id
  route_key = "GET /books/{id}"
  target    = "integrations/${aws_apigatewayv2_integration.get_book_lambda_integration.id}"
  
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito_authorizer.id
}

resource "aws_lambda_permission" "get_book_api_gateway_permission" {
  statement_id  = "AllowAPIGatewayInvokeGetBook"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.get_book_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.books_api.execution_arn}/*/*"
}
