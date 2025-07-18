locals {
  create_book_lambda_source_dir = "${path.module}/lambdas/create-book"
  create_book_go_files_for_hash = fileset(local.create_book_lambda_source_dir, "**/*.go")
  create_book_source_hash       = sha1(join("", [for f in local.create_book_go_files_for_hash : filesha1("${local.create_book_lambda_source_dir}/${f}")]))
}

resource "null_resource" "build_create_book_lambda" {
  triggers = {
    source_hash = local.create_book_source_hash
  }

  provisioner "local-exec" {
    command = "cd ${local.create_book_lambda_source_dir} && make zip"
  }
}

data "aws_iam_policy_document" "create_book_lambda_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "create_book_lambda_exec_role" {
  name               = "create-book-lambda-exec-role"
  assume_role_policy = data.aws_iam_policy_document.create_book_lambda_assume_role_policy.json
}

data "aws_iam_policy_document" "create_book_dynamodb_policy" {
  statement {
    actions   = ["dynamodb:PutItem"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "create_book_dynamodb_policy" {
  name        = "CreateBookDynamoDBPolicy"
  description = "Policy to allow putting an item into the Books DynamoDB table"
  policy      = data.aws_iam_policy_document.create_book_dynamodb_policy.json
}

resource "aws_iam_role_policy_attachment" "create_book_lambda_dynamodb_write" {
  role       = aws_iam_role.create_book_lambda_exec_role.name
  policy_arn = aws_iam_policy.create_book_dynamodb_policy.arn
}

resource "aws_iam_role_policy_attachment" "create_book_lambda_basic_execution" {
  role       = aws_iam_role.create_book_lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_function" "create_book_lambda" {
  function_name = "create-book"
  role          = aws_iam_role.create_book_lambda_exec_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"

  filename         = "${local.create_book_lambda_source_dir}/dist/create-book.zip"
  source_code_hash = local.create_book_source_hash

  depends_on = [
    aws_iam_role_policy_attachment.create_book_lambda_basic_execution,
    aws_iam_role_policy_attachment.create_book_lambda_dynamodb_write,
    null_resource.build_create_book_lambda,
  ]
}

resource "aws_apigatewayv2_integration" "create_book_lambda_integration" {
  api_id           = aws_apigatewayv2_api.books_api.id
  integration_type = "AWS_PROXY"

  integration_uri        = aws_lambda_function.create_book_lambda.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "create_book_route" {
  api_id    = aws_apigatewayv2_api.books_api.id
  route_key = "POST /books"
  target    = "integrations/${aws_apigatewayv2_integration.create_book_lambda_integration.id}"
  
  authorization_type = "JWT"
  authorizer_id      = aws_apigatewayv2_authorizer.cognito_authorizer.id
}

resource "aws_lambda_permission" "create_book_api_gateway_permission" {
  statement_id  = "AllowAPIGatewayInvokeCreateBook"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.create_book_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.books_api.execution_arn}/*/*"
} 