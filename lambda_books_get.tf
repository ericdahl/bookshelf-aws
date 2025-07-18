locals {
  lambda_source_dir = "${path.module}/lambdas/list-books"
  go_files_for_hash = fileset(local.lambda_source_dir, "**/*.go")
  source_hash       = sha1(join("", [for f in local.go_files_for_hash : filesha1("${local.lambda_source_dir}/${f}")]))
}

resource "null_resource" "build_list_books_lambda" {
  triggers = {
    source_hash = local.source_hash
  }

  provisioner "local-exec" {
    command = "cd ${local.lambda_source_dir} && make zip"
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

  filename         = "${local.lambda_source_dir}/dist/list-books.zip"
  source_code_hash = local.source_hash

  environment {
    variables = {
      COGNITO_USER_POOL_ID = aws_cognito_user_pool.bookshelf_user_pool.id
      COGNITO_REGION       = "us-east-1"
    }
  }

  depends_on = [
    aws_iam_role_policy_attachment.list_books_lambda_basic_execution,
    aws_iam_role_policy_attachment.list_books_lambda_dynamodb_read,
    null_resource.build_list_books_lambda,
  ]
} 