# S3 bucket for book exports
resource "aws_s3_bucket" "exports" {
  bucket_prefix = "bookshelf-exports-"
  force_destroy = true
}

# Configure bucket to be private by default
resource "aws_s3_bucket_public_access_block" "exports" {
  bucket = aws_s3_bucket.exports.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

# Lifecycle configuration to automatically delete old exports after 7 days
resource "aws_s3_bucket_lifecycle_configuration" "exports" {
  bucket = aws_s3_bucket.exports.id

  rule {
    id     = "delete_old_exports"
    status = "Enabled"

    expiration {
      days = 7
    }

    noncurrent_version_expiration {
      noncurrent_days = 1
    }
  }
}


