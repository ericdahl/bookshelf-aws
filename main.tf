resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

provider "aws" {
  region = "us-east-1"
  default_tags {
    tags = {
      Name       = "bookshelf-aws"
      Repository = "https://github.com/ericdahl/bookshelf-aws"
    }
  }
}
