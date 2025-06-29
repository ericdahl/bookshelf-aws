locals {
  get_book_lambda_source_dir = "${path.module}/lambdas/get-book"
  get_book_go_files_for_hash = fileset(local.get_book_lambda_source_dir, "**/*.go")
  get_book_source_hash       = sha1(join("", [for f in local.get_book_go_files_for_hash : filesha1("${local.get_book_lambda_source_dir}/${f}")]))
}

data "external" "build_get_book_lambda" {
  program = ["bash", "-c", <<EOT
set -e
SOURCE_DIR="${local.get_book_lambda_source_dir}"
OUTPUT_DIR=$(mktemp -d)
ZIP_PATH="$OUTPUT_DIR/lambda.zip"

cd "$SOURCE_DIR"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$OUTPUT_DIR/bootstrap" main.go

cd "$OUTPUT_DIR"
zip "$ZIP_PATH" bootstrap > /dev/null

jq -n --arg filename "$ZIP_PATH" '{"filename": $filename}'
EOT
  ]

  query = {
    source_hash = local.get_book_source_hash
  }
}

resource "aws_lambda_function" "get_book_lambda" {
  function_name = "get-book"
  role          = aws_iam_role.books_lambda_exec_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"

  filename         = data.external.build_get_book_lambda.result.filename
  source_code_hash = filebase64sha256(data.external.build_get_book_lambda.result.filename)

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic_execution,
    aws_iam_role_policy_attachment.lambda_dynamodb_read,
  ]
}

resource "aws_apigatewayv2_integration" "get_book_lambda_integration" {
  api_id           = aws_apigatewayv2_api.books_api.id
  integration_type = "AWS_PROXY"

  integration_uri    = aws_lambda_function.get_book_lambda.invoke_arn
  payload_format_version = "2.0"
}

resource "aws_apigatewayv2_route" "get_book_route" {
  api_id    = aws_apigatewayv2_api.books_api.id
  route_key = "GET /books/{id}"
  target    = "integrations/${aws_apigatewayv2_integration.get_book_lambda_integration.id}"
}

resource "aws_lambda_permission" "get_book_api_gateway_permission" {
  statement_id  = "AllowAPIGatewayInvokeGetBook"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.get_book_lambda.function_name
  principal     = "apigateway.amazonaws.com"

  source_arn = "${aws_apigatewayv2_api.books_api.execution_arn}/*/*"
}
