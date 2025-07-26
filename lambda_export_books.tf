# Build the export-books Lambda function
resource "null_resource" "export_books_build" {
  provisioner "local-exec" {
    command     = "make build"
    working_dir = "${path.module}/lambdas/export-books"
  }

  triggers = {
    # Rebuild when source files change
    source_hash = filesha256("${path.module}/lambdas/export-books/main.go")
    go_mod_hash = filesha256("${path.module}/lambdas/export-books/go.mod")
  }
}

# Create the Lambda function for exporting books
resource "aws_lambda_function" "export_books" {
  depends_on = [null_resource.export_books_build]

  filename         = "lambdas/export-books/function.zip"
  function_name    = "bookshelf-export-books"
  role            = aws_iam_role.export_books_lambda.arn
  handler         = "bootstrap"
  runtime         = "provided.al2"
  timeout         = 30
  memory_size     = 256

  environment {
    variables = {
      EXPORTS_BUCKET_NAME = aws_s3_bucket.exports.bucket
    }
  }

  source_code_hash = data.archive_file.export_books_zip.output_base64sha256
}

# Archive the Lambda function code
data "archive_file" "export_books_zip" {
  depends_on  = [null_resource.export_books_build]
  type        = "zip"
  source_file = "lambdas/export-books/dist/bootstrap"
  output_path = "lambdas/export-books/function.zip"
}

# IAM role for the export-books Lambda function
resource "aws_iam_role" "export_books_lambda" {
  name = "bookshelf-export-books-lambda-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

# Attach basic execution role
resource "aws_iam_role_policy_attachment" "export_books_lambda_basic" {
  role       = aws_iam_role.export_books_lambda.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

# IAM policy for DynamoDB access
resource "aws_iam_role_policy" "export_books_dynamodb" {
  name = "bookshelf-export-books-dynamodb-policy"
  role = aws_iam_role.export_books_lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "dynamodb:Query"
        ]
        Resource = [
          aws_dynamodb_table.books.arn,
          "${aws_dynamodb_table.books.arn}/index/*"
        ]
      }
    ]
  })
}

# IAM policy for S3 access to exports bucket
resource "aws_iam_role_policy" "export_books_s3" {
  name = "bookshelf-export-books-s3-policy"
  role = aws_iam_role.export_books_lambda.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "s3:PutObject",
          "s3:GetObject",
          "s3:DeleteObject"
        ]
        Resource = "${aws_s3_bucket.exports.arn}/*"
      }
    ]
  })
}