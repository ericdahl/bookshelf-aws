locals {
  lambda_source_dir = "${path.module}/lambdas/list-books"
  go_files_for_hash = fileset(local.lambda_source_dir, "**/*.go")
  source_hash       = sha1(join("", [for f in local.go_files_for_hash : filesha1("${local.lambda_source_dir}/${f}")]))
}

data "external" "build_list_books_lambda" {
  program = ["bash", "-c", <<EOT
set -e
SOURCE_DIR="${local.lambda_source_dir}"
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
    # This query makes sure the external data source is re-triggered when .go files change
    source_hash = local.source_hash
  }
}

data "aws_iam_policy_document" "books_lambda_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "books_lambda_exec_role" {
  name               = "books-lambda-exec-role"
  assume_role_policy = data.aws_iam_policy_document.books_lambda_assume_role_policy.json
}

data "aws_iam_policy_document" "dynamodb_read_policy" {
  statement {
    actions   = ["dynamodb:Scan", "dynamodb:GetItem"]
    resources = ["*"]
  }
}

resource "aws_iam_policy" "dynamodb_read_policy" {
  name        = "DynamoDBReadPolicy"
  description = "Policy to allow reading from the Books DynamoDB table"
  policy      = data.aws_iam_policy_document.dynamodb_read_policy.json
}

resource "aws_iam_role_policy_attachment" "lambda_dynamodb_read" {
  role       = aws_iam_role.books_lambda_exec_role.name
  policy_arn = aws_iam_policy.dynamodb_read_policy.arn
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.books_lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_function" "list_books_lambda" {
  function_name = "list-books"
  role          = aws_iam_role.books_lambda_exec_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"

  filename         = data.external.build_list_books_lambda.result.filename
  source_code_hash = filebase64sha256(data.external.build_list_books_lambda.result.filename)

  depends_on = [
    aws_iam_role_policy_attachment.lambda_basic_execution,
    aws_iam_role_policy_attachment.lambda_dynamodb_read,
  ]
} 