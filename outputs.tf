output "api_gateway_invoke_url" {
  description = "The invoke URL for the API Gateway stage"
  value       = aws_apigatewayv2_stage.api_stage.invoke_url
}

output "website_url" {
  description = "The URL of the static website hosted on S3"
  value       = aws_s3_bucket_website_configuration.bookshelf_web.website_endpoint
}

output "s3_bucket_name" {
  description = "The name of the S3 bucket hosting the website"
  value       = aws_s3_bucket.bookshelf_web.bucket
}
