
provider "aws" {
  region = "us-east-1"
  default_tags {
    tags = {
      Name       = "bookshelf-aws"
      Repository = "https://github.com/ericdahl/bookshelf-aws"
    }
  }
}
