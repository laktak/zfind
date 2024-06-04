
# zfind

zfind allows you to search for files, including inside `tar` and `zip` archives. It makes finding files easy with a filter syntax that is similar to an SQL-WHERE clause.

## Basic Usage

```
zfind -w <where> <path>
```

Examples

```
# find files smaller than 10KB
zfind -w 'size<10k'

# find files modified before 2010 inside a tar
zfind -w 'date<"2010" and archive="tar"'

# find files named *.go and modified today
zfind -w 'name like "%.go" and date=today'

# find directories named foo and bar
zfind -w 'name in ("foo", "bar") and type="dir"'

# search for all README.md files and show in long listing format
zfind -w 'name="README.md"' -l

# show results in csv format
zfind --csv
```

## Where

- `AND`, `OR` and `()` parentheses are logical operators used to combine multiple conditions. 'AND' means that both conditions must be true for a row to be included in the results. 'OR' means that if either condition is true, the row will be included. Parentheses are used to group conditions, just like in mathematics.

Example: `-w '(size > 20M OR name = "temp") AND type="file"'` selects all files that are either greater than 20MB in size or are named temp.

- Operators `=`, `<>`, `!=`, `<`, `>`, `<=`, `>=` are comparison operators used to compare values and file properties. The types must match, meaning don't compare a date to a file size.

Example: `-w 'date > "2020-10-01"` selects all files that were modified after the specified date.

- `LIKE`, `ILIKE` and `RLIKE` are used for pattern matching in strings.
  - `LIKE` is case-sensitive, while `ILIKE` is case-insensitive.
  - The `%` symbol is used as a wildcard character that matches any sequence of characters.
  - The `_` symbol matches any single character.
  - `RLIKE` allows to match a regular expression.

Example: `-w "name like "z%"` selects all files whose name starts with 'z'.

- `IN` allows you to specify multiple values to match. A file will be included if the value of the property matches any of the values in the list.

Example: `-w "type in ("file", "link")` selects all files of type file or link.

- `BETWEEN` selects values within a given range (inclusive).

Example: `-w "date between "2010" and "2011-01-15"` means that all files that were modified from 2010 to 2011-01-15 will be included.

- `NOT` is a logical operator used to negate a condition. It returns true if the condition is false and vice versa.

Example: `-w "name not like "z%"`, `-w "date not between "2010" and "2011-01-15"`, `-w "type not in ("file", "link")`

- Values can be numbers, text, date and time, `TRUE` and `FALSE`
  - dates have to be specified in the YYYY-MM-DD format
  - times have to be specified in the 24h HH:MM:SS format
  - numbers can be written as sizes by appending `B`, `K`, `M`, `G` and `T` to specify bytes, KB, MB, GB and TB.
  - `TRUE` and `FALSE`:
  - empty strings and `0` evaluate to `false`

## Properties

The following file properies are available:

| name        | description                                                    |
|-------------|----------------------------------------------------------------|
| name        | name of the file                                               |
| path        | full path of the file                                          |
| container   | path of the container (if inside a zip or tar)                 |
| size        | file size (uncompressed)                                       |
| date        | modified date in YYYY-MM-DD format                             |
| time        | modified time in HH-MM-SS format                               |
| type        | `file`, `dir`, or `link`                                       |
| archive     | archive type: `tar`, `zip` or empty                            |

Helper properties

| name        | description                                                    |
|-------------|----------------------------------------------------------------|
| today       | todays date                                                    |


## Installation

zfind is built for a number of platforms by GitHub actions.

Download a binary from [releases](https://github.com/laktak/zfind/releases) and place it in your `PATH`.

