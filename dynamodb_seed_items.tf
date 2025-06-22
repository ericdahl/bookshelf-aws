resource "aws_dynamodb_table_item" "the_way_of_kings" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#THE_WAY_OF_KINGS" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "The Way of Kings" }
    Author = { "S" = "Brandon Sanderson" }
    Series = { "S" = "The Stormlight Archive" }
    status = { "S" = "READ" }
  })
}
