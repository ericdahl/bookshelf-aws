locals {
  search_lambda_source_dir = "${path.module}/lambdas/search-books"
  search_go_files_for_hash = fileset(local.search_lambda_source_dir, "**/*.go")
  search_source_hash       = sha1(join("", [for f in local.search_go_files_for_hash : filesha1("${local.search_lambda_source_dir}/${f}")]))
}

resource "null_resource" "build_search_books_lambda" {
  triggers = {
    source_hash = local.search_source_hash
  }

  provisioner "local-exec" {
    command = "cd ${local.search_lambda_source_dir} && make zip"
  }
}

data "aws_iam_policy_document" "search_books_lambda_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "search_books_lambda_exec_role" {
  name               = "search-books-lambda-exec-role"
  assume_role_policy = data.aws_iam_policy_document.search_books_lambda_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "search_books_lambda_basic_execution" {
  role       = aws_iam_role.search_books_lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_function" "search_books_lambda" {
  function_name = "search-books"
  role          = aws_iam_role.search_books_lambda_exec_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"

  filename         = "${local.search_lambda_source_dir}/dist/search-books.zip"
  source_code_hash = local.search_source_hash

  depends_on = [
    aws_iam_role_policy_attachment.search_books_lambda_basic_execution,
    null_resource.build_search_books_lambda,
  ]
}

resource "aws_apigatewayv2_integration" "search_lambda_integration" {
  api_id           = aws_apigatewayv2_api.books_api.id
  integration_type = "AWS_PROXY"

  integration_uri        = aws_lambda_function.search_books_lambda.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "search_books_route" {
  api_id    = aws_apigatewayv2_api.books_api.id
  route_key = "GET /search"
  target    = "integrations/${aws_apigatewayv2_integration.search_lambda_integration.id}"
}

resource "aws_lambda_permission" "search_api_gateway_permission" {
  statement_id  = "AllowAPIGatewayInvokeSearch"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.search_books_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.books_api.execution_arn}/*/*"
}