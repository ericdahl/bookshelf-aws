resource "aws_dynamodb_table" "books" {
  name         = "books"
  billing_mode = "PAY_PER_REQUEST"
  hash_key     = "PK"
  range_key    = "SK"

  attribute {
    name = "PK"
    type = "S"
  }

  attribute {
    name = "SK"
    type = "S"
  }

  attribute {
    name = "status"
    type = "S"
  }

  global_secondary_index {
    name            = "GSI1"
    hash_key        = "status"
    range_key       = "PK"
    projection_type = "ALL"
  }
} 