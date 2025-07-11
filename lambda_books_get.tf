locals {
  lambda_source_dir = "${path.module}/lambdas/list-books"
  go_files_for_hash = fileset(local.lambda_source_dir, "**/*.go")
  source_hash       = sha1(join("", [for f in local.go_files_for_hash : filesha1("${local.lambda_source_dir}/${f}")]))
}

data "external" "build_list_books_lambda" {
  program = ["bash", "-c", <<EOT
set -e
SOURCE_DIR="${local.lambda_source_dir}"
cd "$SOURCE_DIR"
make zip >&2
jq -n --arg filename "$SOURCE_DIR/dist/list-books.zip" '{"filename": $filename}'
EOT
  ]

  query = {
    # This query makes sure the external data source is re-triggered when .go files change
    source_hash = local.source_hash
  }
}

data "aws_iam_policy_document" "list_books_lambda_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "list_books_lambda_exec_role" {
  name               = "list-books-lambda-exec-role"
  assume_role_policy = data.aws_iam_policy_document.list_books_lambda_assume_role_policy.json
}

data "aws_iam_policy_document" "list_books_dynamodb_policy" {
  statement {
    actions = [
      "dynamodb:Scan",
      "dynamodb:Query"
    ]
    resources = [
      aws_dynamodb_table.books.arn,
      "${aws_dynamodb_table.books.arn}/index/*",
    ]
  }
}

resource "aws_iam_policy" "list_books_dynamodb_policy" {
  name        = "ListBooksDynamoDBPolicy"
  description = "Policy to allow scanning the Books DynamoDB table"
  policy      = data.aws_iam_policy_document.list_books_dynamodb_policy.json
}

resource "aws_iam_role_policy_attachment" "list_books_lambda_dynamodb_read" {
  role       = aws_iam_role.list_books_lambda_exec_role.name
  policy_arn = aws_iam_policy.list_books_dynamodb_policy.arn
}

resource "aws_iam_role_policy_attachment" "list_books_lambda_basic_execution" {
  role       = aws_iam_role.list_books_lambda_exec_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_lambda_function" "list_books_lambda" {
  function_name = "list-books"
  role          = aws_iam_role.list_books_lambda_exec_role.arn
  handler       = "bootstrap"
  runtime       = "provided.al2"

  filename         = data.external.build_list_books_lambda.result.filename
  source_code_hash = filebase64sha256(data.external.build_list_books_lambda.result.filename)

  depends_on = [
    aws_iam_role_policy_attachment.list_books_lambda_basic_execution,
    aws_iam_role_policy_attachment.list_books_lambda_dynamodb_read,
  ]
} 