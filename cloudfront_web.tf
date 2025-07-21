# CloudFront Origin Access Control for S3
resource "aws_cloudfront_origin_access_control" "bookshelf_oac" {
  name                              = "bookshelf-s3-oac"
  description                       = "Origin Access Control for Bookshelf S3 bucket"
  origin_access_control_origin_type = "s3"
  signing_behavior                  = "always"
  signing_protocol                  = "sigv4"
}

# CloudFront function to rewrite /api paths
resource "aws_cloudfront_function" "api_path_rewrite" {
  name    = "api-path-rewrite"
  runtime = "cloudfront-js-2.0"
  comment = "Rewrite /api paths to remove /api prefix for API Gateway"
  publish = true
  code    = <<-EOT
async function handler(event) {
    var request = event.request;
    var uri = request.uri;
    
    // If the URI starts with /api/, remove the /api prefix
    if (uri.startsWith('/api/')) {
        request.uri = uri.substring(4); // Remove '/api' (4 characters)
    }
    
    return request;
}
EOT
}

# CloudFront distribution
resource "aws_cloudfront_distribution" "bookshelf_web" {
  enabled             = true
  is_ipv6_enabled     = true
  default_root_object = "index.html"
  comment             = "Bookshelf web application distribution"

  # S3 origin for static content
  origin {
    domain_name              = aws_s3_bucket.bookshelf_web.bucket_regional_domain_name
    origin_id                = "S3-${aws_s3_bucket.bookshelf_web.bucket}"
    origin_access_control_id = aws_cloudfront_origin_access_control.bookshelf_oac.id
  }

  # API Gateway origin for API calls
  origin {
    domain_name = replace(aws_apigatewayv2_api.books_api.api_endpoint, "https://", "")
    origin_id   = "API-Gateway"
    origin_path = "/prod"

    custom_origin_config {
      http_port              = 443
      https_port             = 443
      origin_protocol_policy = "https-only"
      origin_ssl_protocols   = ["TLSv1.2"]
    }
  }

  # Default cache behavior for static content (S3)
  default_cache_behavior {
    allowed_methods        = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods         = ["GET", "HEAD"]
    target_origin_id       = "S3-${aws_s3_bucket.bookshelf_web.bucket}"
    compress               = true
    viewer_protocol_policy = "redirect-to-https"
    cache_policy_id        = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # CachingOptimized - respects origin cache headers

    # Remove forwarded_values, min_ttl, default_ttl, max_ttl when using cache_policy_id
  }

  # Cache behavior for API routes
  ordered_cache_behavior {
    path_pattern             = "/api/*"
    allowed_methods          = ["DELETE", "GET", "HEAD", "OPTIONS", "PATCH", "POST", "PUT"]
    cached_methods           = ["GET", "HEAD", "OPTIONS"]
    target_origin_id         = "API-Gateway"
    compress                 = true
    viewer_protocol_policy   = "https-only"
    cache_policy_id          = "4135ea2d-6df8-44a3-9df3-4b5a84be39ad" # CachingOptimized - respects origin cache headers
    origin_request_policy_id = "b689b0a8-53d0-40ab-baf2-68738e2966ac" # AllViewerExceptHostHeader - forwards all headers except Host

    function_association {
      event_type   = "viewer-request"
      function_arn = aws_cloudfront_function.api_path_rewrite.arn
    }
  }

  # Geographic restrictions
  restrictions {
    geo_restriction {
      restriction_type = "none"
    }
  }

  # SSL certificate
  viewer_certificate {
    cloudfront_default_certificate = true
  }


  tags = {
    Name = "bookshelf-web-distribution"
  }

  depends_on = [aws_apigatewayv2_api.books_api]
}

# Output the CloudFront domain name
output "cloudfront_domain_name" {
  value       = aws_cloudfront_distribution.bookshelf_web.domain_name
  description = "CloudFront distribution domain name"
}

output "cloudfront_distribution_id" {
  value       = aws_cloudfront_distribution.bookshelf_web.id
  description = "CloudFront distribution ID"
}