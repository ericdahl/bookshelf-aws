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

resource "aws_dynamodb_table_item" "words_of_radiance" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#WORDS_OF_RADIANCE" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "Words of Radiance" }
    Author = { "S" = "Brandon Sanderson" }
    Series = { "S" = "The Stormlight Archive" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "oathbringer" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#OATHBRINGER" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "Oathbringer" }
    Author = { "S" = "Brandon Sanderson" }
    Series = { "S" = "The Stormlight Archive" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "rhythm_of_war" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#RHYTHM_OF_WAR" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "Rhythm of War" }
    Author = { "S" = "Brandon Sanderson" }
    Series = { "S" = "The Stormlight Archive" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "wind_and_truth" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#WIND_AND_TRUTH" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "Wind and Truth" }
    Author = { "S" = "Brandon Sanderson" }
    Series = { "S" = "The Stormlight Archive" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "the_eye_of_the_world" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#THE_EYE_OF_THE_WORLD" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "The Eye of the World" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "the_great_hunt" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#THE_GREAT_HUNT" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "The Great Hunt" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "the_dragon_reborn" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#THE_DRAGON_REBORN" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "The Dragon Reborn" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "the_shadow_rising" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#THE_SHADOW_RISING" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "The Shadow Rising" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "the_fires_of_heaven" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#THE_FIRES_OF_HEAVEN" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "The Fires of Heaven" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "lord_of_chaos" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#LORD_OF_CHAOS" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "Lord of Chaos" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "a_crown_of_swords" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#A_CROWN_OF_SWORDS" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "A Crown of Swords" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "the_path_of_daggers" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#THE_PATH_OF_DAGGERS" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "The Path of Daggers" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "winters_heart" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#WINTERS_HEART" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "Winter's Heart" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "crossroads_of_twilight" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#CROSSROADS_OF_TWILIGHT" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "Crossroads of Twilight" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "knife_of_dreams" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#KNIFE_OF_DREAMS" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "Knife of Dreams" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "the_gathering_storm" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#THE_GATHERING_STORM" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "The Gathering Storm" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "towers_of_midnight" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#TOWERS_OF_MIDNIGHT" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "Towers of Midnight" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}

resource "aws_dynamodb_table_item" "a_memory_of_light" {
  table_name = aws_dynamodb_table.books.name
  hash_key   = aws_dynamodb_table.books.hash_key
  range_key  = aws_dynamodb_table.books.range_key

  item = jsonencode({
    PK     = { "S" = "BOOK#A_MEMORY_OF_LIGHT" }
    SK     = { "S" = "BOOK" }
    Title  = { "S" = "A Memory of Light" }
    Author = { "S" = "Robert Jordan" }
    Series = { "S" = "The Wheel of Time" }
    status = { "S" = "WANT_TO_READ" }
  })
}
