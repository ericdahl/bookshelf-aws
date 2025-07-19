resource "aws_dynamodb_table_item" "the_way_of_kings" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK        = { "S" = "BOOK#a1b2c3d4-e5f6-7890-1234-567890abcdef" }
    SK        = { "S" = "BOOK" }
    id        = { "S" = "a1b2c3d4-e5f6-7890-1234-567890abcdef" }
    Title     = { "S" = "The Way of Kings" }
    Author    = { "S" = "Brandon Sanderson" }
    Series    = { "S" = "The Stormlight Archive" }
    status    = { "S" = "READ" }
    thumbnail = { "S" = "https://books.google.com/books/content?id=X_x_AAAAQBAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api" }
  })
}

resource "aws_dynamodb_table_item" "words_of_radiance" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK        = { "S" = "BOOK#b2c3d4e5-f6a7-8901-2345-67890abcdef1" }
    SK        = { "S" = "BOOK" }
    id        = { "S" = "b2c3d4e5-f6a7-8901-2345-67890abcdef1" }
    Title     = { "S" = "Words of Radiance" }
    Author    = { "S" = "Brandon Sanderson" }
    Series    = { "S" = "The Stormlight Archive" }
    status    = { "S" = "WANT_TO_READ" }
    thumbnail = { "S" = "https://books.google.com/books/content?id=kYjqAQAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api" }
  })
}

resource "aws_dynamodb_table_item" "oathbringer" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK        = { "S" = "BOOK#c3d4e5f6-a7b8-9012-3456-7890abcdef12" }
    SK        = { "S" = "BOOK" }
    id        = { "S" = "c3d4e5f6-a7b8-9012-3456-7890abcdef12" }
    Title     = { "S" = "Oathbringer" }
    Author    = { "S" = "Brandon Sanderson" }
    Series    = { "S" = "The Stormlight Archive" }
    status    = { "S" = "WANT_TO_READ" }
    thumbnail = { "S" = "https://books.google.com/books/content?id=VsT3DQAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api" }
  })
}

resource "aws_dynamodb_table_item" "rhythm_of_war" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK        = { "S" = "BOOK#d4e5f6a7-b8c9-0123-4567-890abcdef123" }
    SK        = { "S" = "BOOK" }
    id        = { "S" = "d4e5f6a7-b8c9-0123-4567-890abcdef123" }
    Title     = { "S" = "Rhythm of War" }
    Author    = { "S" = "Brandon Sanderson" }
    Series    = { "S" = "The Stormlight Archive" }
    status    = { "S" = "WANT_TO_READ" }
    thumbnail = { "S" = "https://books.google.com/books/content?id=QCPBDwAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api" }
  })
}

resource "aws_dynamodb_table_item" "wind_and_truth" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK        = { "S" = "BOOK#e5f6a7b8-c9d0-1234-5678-90abcdef1234" }
    SK        = { "S" = "BOOK" }
    id        = { "S" = "e5f6a7b8-c9d0-1234-5678-90abcdef1234" }
    Title     = { "S" = "Wind and Truth" }
    Author    = { "S" = "Brandon Sanderson" }
    Series    = { "S" = "The Stormlight Archive" }
    status    = { "S" = "WANT_TO_READ" }
    thumbnail = { "S" = "https://books.google.com/books/content?id=GInoEAAAQBAJ&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api" }
  })
}

resource "aws_dynamodb_table_item" "the_eye_of_the_world" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK        = { "S" = "BOOK#f6a7b8c9-d0e1-2345-6789-0abcdef12345" }
    SK        = { "S" = "BOOK" }
    id        = { "S" = "f6a7b8c9-d0e1-2345-6789-0abcdef12345" }
    Title     = { "S" = "The Eye of the World" }
    Author    = { "S" = "Robert Jordan" }
    Series    = { "S" = "The Wheel of Time" }
    status    = { "S" = "WANT_TO_READ" }
    thumbnail = { "S" = "https://books.google.com/books/content?id=PmJuDwAAQBAJ&printsec=frontcover&img=1&zoom=1&source=gbs_api" }
  })
}

resource "aws_dynamodb_table_item" "the_great_hunt" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK        = { "S" = "BOOK#a7b8c9d0-e1f2-3456-7890-bcdef123456" }
    SK        = { "S" = "BOOK" }
    id        = { "S" = "a7b8c9d0-e1f2-3456-7890-bcdef123456" }
    Title     = { "S" = "The Great Hunt" }
    Author    = { "S" = "Robert Jordan" }
    Series    = { "S" = "The Wheel of Time" }
    status    = { "S" = "WANT_TO_READ" }
    thumbnail = { "S" = "https://books.google.com/books/content?id=yngEsxEO4QYC&printsec=frontcover&img=1&zoom=1&edge=curl&source=gbs_api" }
  })
}
