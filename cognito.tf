resource "aws_cognito_user_pool" "bookshelf_user_pool" {
  name = "bookshelf-users"

  # Password policy
  password_policy {
    minimum_length                   = 8
    require_lowercase                = true
    require_numbers                  = true
    require_symbols                  = true
    require_uppercase                = true
    temporary_password_validity_days = 7
  }

  # User attributes
  username_attributes = ["email"]
  
  # Auto verification
  auto_verified_attributes = ["email"]

  # Email configuration
  email_configuration {
    email_sending_account = "COGNITO_DEFAULT"
  }

  # Account recovery
  account_recovery_setting {
    recovery_mechanism {
      name     = "verified_email"
      priority = 1
    }
  }

  # User pool add-ons
  user_pool_add_ons {
    advanced_security_mode = "ENFORCED"
  }

  # Verification message templates
  verification_message_template {
    default_email_option = "CONFIRM_WITH_CODE"
    email_message        = "Your verification code is {####}"
    email_subject        = "Verify your Bookshelf account"
  }

  tags = {
    Name = "bookshelf-user-pool"
  }
}

resource "aws_cognito_user_pool_client" "bookshelf_client" {
  name         = "bookshelf-client"
  user_pool_id = aws_cognito_user_pool.bookshelf_user_pool.id

  # Client settings
  generate_secret                      = false
  prevent_user_existence_errors        = "ENABLED"
  enable_token_revocation              = true
  enable_propagate_additional_user_context_data = false

  # Auth flows
  explicit_auth_flows = [
    "ALLOW_USER_PASSWORD_AUTH",
    "ALLOW_REFRESH_TOKEN_AUTH",
    "ALLOW_USER_SRP_AUTH"
  ]

  # Token validity
  access_token_validity  = 24
  id_token_validity     = 24
  refresh_token_validity = 30

  token_validity_units {
    access_token  = "hours"
    id_token      = "hours"
    refresh_token = "days"
  }

  # Read and write attributes
  read_attributes = [
    "email",
    "email_verified",
    "name",
    "preferred_username"
  ]

  write_attributes = [
    "email",
    "name",
    "preferred_username"
  ]
}

# Outputs for later use
output "cognito_user_pool_id" {
  description = "ID of the Cognito User Pool"
  value       = aws_cognito_user_pool.bookshelf_user_pool.id
}

output "cognito_user_pool_arn" {
  description = "ARN of the Cognito User Pool"
  value       = aws_cognito_user_pool.bookshelf_user_pool.arn
}

output "cognito_user_pool_client_id" {
  description = "ID of the Cognito User Pool Client"
  value       = aws_cognito_user_pool_client.bookshelf_client.id
}

output "cognito_user_pool_domain" {
  description = "Domain of the Cognito User Pool"
  value       = aws_cognito_user_pool.bookshelf_user_pool.domain
}

# Test user for automated testing
resource "aws_cognito_user" "test_user" {
  user_pool_id = aws_cognito_user_pool.bookshelf_user_pool.id
  username     = "testuser@example.com"
  
  attributes = {
    email           = "testuser@example.com"
    email_verified  = "true"
    name           = "Test User"
  }
  
  password = "TestPassword123!"
  
  # Ensure the user is confirmed and doesn't need email verification
  message_action = "SUPPRESS"
}

# Output test user credentials for automated tests
output "test_user_email" {
  description = "Email of the test user for automated testing"
  value       = aws_cognito_user.test_user.username
  sensitive   = true
}

output "test_user_password" {
  description = "Password of the test user for automated testing"
  value       = aws_cognito_user.test_user.password
  sensitive   = true
}