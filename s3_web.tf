# S3 bucket for static website hosting
resource "aws_s3_bucket" "bookshelf_web" {
  bucket = "bookshelf-web-${random_id.bucket_suffix.hex}"
}

# Random ID for bucket naming
resource "random_id" "bucket_suffix" {
  byte_length = 8
}

# S3 bucket website configuration
resource "aws_s3_bucket_website_configuration" "bookshelf_web" {
  bucket = aws_s3_bucket.bookshelf_web.id

  index_document {
    suffix = "index.html"
  }

  error_document {
    key = "404.html"
  }
}

# S3 bucket public access block - secure configuration
resource "aws_s3_bucket_public_access_block" "bookshelf_web" {
  bucket = aws_s3_bucket.bookshelf_web.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# S3 bucket policy for CloudFront access via OAC
resource "aws_s3_bucket_policy" "bookshelf_web" {
  bucket     = aws_s3_bucket.bookshelf_web.id
  depends_on = [aws_s3_bucket_public_access_block.bookshelf_web]

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowCloudFrontServicePrincipal"
        Effect = "Allow"
        Principal = {
          Service = "cloudfront.amazonaws.com"
        }
        Action   = "s3:GetObject"
        Resource = "${aws_s3_bucket.bookshelf_web.arn}/*"
        Condition = {
          StringEquals = {
            "AWS:SourceArn" = aws_cloudfront_distribution.bookshelf_web.arn
          }
        }
      },
    ]
  })
}

# Upload index.html
resource "aws_s3_object" "index_html" {
  bucket       = aws_s3_bucket.bookshelf_web.id
  key          = "index.html"
  source       = "web/index.html"
  content_type = "text/html"
  etag         = filemd5("web/index.html")
}

# Upload CSS files
resource "aws_s3_object" "styles_css" {
  bucket       = aws_s3_bucket.bookshelf_web.id
  key          = "css/styles.css"
  source       = "web/css/styles.css"
  content_type = "text/css"
  etag         = filemd5("web/css/styles.css")
}

# Upload JavaScript files
resource "aws_s3_object" "app_js" {
  bucket       = aws_s3_bucket.bookshelf_web.id
  key          = "js/app.js"
  source       = "web/js/app.js"
  content_type = "application/javascript"
  etag         = filemd5("web/js/app.js")
}

resource "aws_s3_object" "config_js" {
  bucket       = aws_s3_bucket.bookshelf_web.id
  key          = "js/config.js"
  source       = "web/js/config.js"
  content_type = "application/javascript"
  etag         = filemd5("web/js/config.js")
}

# Upload favicon
resource "aws_s3_object" "favicon" {
  bucket       = aws_s3_bucket.bookshelf_web.id
  key          = "favicon.ico"
  source       = "web/favicon.ico"
  content_type = "image/x-icon"
  etag         = filemd5("web/favicon.ico")
}

# Upload icon
resource "aws_s3_object" "icon_png" {
  bucket       = aws_s3_bucket.bookshelf_web.id
  key          = "icon.png"
  source       = "web/icon.png"
  content_type = "image/png"
  etag         = filemd5("web/icon.png")
}

# Upload authentication pages
resource "aws_s3_object" "signup_html" {
  bucket       = aws_s3_bucket.bookshelf_web.id
  key          = "signup.html"
  source       = "web/signup.html"
  content_type = "text/html"
  etag         = filemd5("web/signup.html")
}

resource "aws_s3_object" "signin_html" {
  bucket       = aws_s3_bucket.bookshelf_web.id
  key          = "signin.html"
  source       = "web/signin.html"
  content_type = "text/html"
  etag         = filemd5("web/signin.html")
}

resource "aws_s3_object" "verify_html" {
  bucket       = aws_s3_bucket.bookshelf_web.id
  key          = "verify.html"
  source       = "web/verify.html"
  content_type = "text/html"
  etag         = filemd5("web/verify.html")
}

resource "aws_s3_object" "error_404_html" {
  bucket       = aws_s3_bucket.bookshelf_web.id
  key          = "404.html"
  source       = "web/404.html"
  content_type = "text/html"
  etag         = filemd5("web/404.html")
}

# Upload profile page
resource "aws_s3_object" "profile_html" {
  bucket       = aws_s3_bucket.bookshelf_web.id
  key          = "profile.html"
  source       = "web/profile.html"
  content_type = "text/html"
  etag         = filemd5("web/profile.html")
}

# Upload profile JavaScript
resource "aws_s3_object" "profile_js" {
  bucket       = aws_s3_bucket.bookshelf_web.id
  key          = "js/profile.js"
  source       = "web/js/profile.js"
  content_type = "application/javascript"
  etag         = filemd5("web/js/profile.js")
} 