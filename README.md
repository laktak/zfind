
# zfind

`zfind` allows you to search for files, including inside `tar`, `zip`, `7z` and `rar` archives. It makes finding files easy with a filter syntax that is similar to an SQL-WHERE clause. This means, if you know SQL, you don't have to learn or remember any new syntax just for this tool.

- [Basic Usage & Examples](#basic-usage--examples)
- [Where Syntax](#where-syntax)
- [Properties](#properties)
- [Supported archives](#supported-archives)
- [Actions](#actions)
- [Configuration](#configuration)
- [Installation](#installation)
  - [Binary releases](#binary-releases)
  - [Homebrew (macOS and Linux)](#homebrew-macos-and-linux)
  - [Arch Linux](#arch-linux)
  - [Build from Source](#build-from-source)
- [zfind as a Go module](#zfind-as-a-go-module)


## Basic Usage & Examples

```shell
zfind <where> [<path>...]
```

Examples

```console
# find files smaller than 10KB, in the current path
zfind 'size<10k'

# find files in the given range in /some/path
zfind 'size between 1M and 1G' /some/path

# find files modified before 2010 inside a tar
zfind 'date<"2010" and archive="tar"'

# find files named foo* and modified today
zfind 'name like "foo%" and date=today'

# find files that contain two dashes using a regex
zfind 'name rlike "(.*-){2}"'

# find files that have the extension .jpg or .jpeg
zfind 'ext in ("jpg","jpeg")'

# find directories named foo and bar
zfind 'name in ("foo", "bar") and type="dir"'

# search for all README.md files and show in long listing format
zfind 'name="README.md"' -l

# show results in csv format
zfind --csv
zfind --csv-no-head
```

## Where Syntax

- `AND`, `OR` and `()` parentheses are logical operators used to combine multiple conditions. `AND` means that both conditions must be true for a row to be included in the results. `OR` means that if either condition is true, the row will be included. Parentheses are used to group conditions, just like in mathematics.

Example: `'(size > 20M OR name = "temp") AND type="file"'` selects all files that are either greater than 20 MB in size or are named temp.

- Operators `=`, `<>`, `!=`, `<`, `>`, `<=`, `>=` are comparison operators used to compare values and file properties. The types must match, meaning don't compare a date to a file size.

Example: `'date > "2020-10-01"'` selects all files that were modified after the specified date.

- `LIKE`, `ILIKE` and `RLIKE` are used for pattern matching in strings.
  - `LIKE` is case-sensitive, while `ILIKE` is case-insensitive.
  - The `%` symbol is used as a wildcard character that matches any sequence of characters.
  - The `_` symbol matches any single character.
  - `RLIKE` allows matching a regular expression.

Example: `'"name like "z%"'` selects all files whose name starts with 'z'.

- `IN` allows you to specify multiple values to match. A file will be included if the value of the property matches any of the values in the list.

Example: `'"type in ("file", "link")'` selects all files of type file or link.

- `BETWEEN` selects values within a given range (inclusive).

Example: `'"date between "2010" and "2011-01-15"'` means that all files that were modified from 2010 to 2011-01-15 will be included.

- `NOT` is a logical operator used to negate a condition. It returns true if the condition is false and vice versa.

Example: `'"name not like "z%"'`, `'"date not between "2010" and "2011-01-15"'`, `'"type not in ("file", "link")'`

- Values can be numbers, text, date and time, `TRUE` and `FALSE`
  - dates have to be specified in `YYYY-MM-DD` format
  - times have to be specified in 24h `HH:MM:SS` format
  - numbers can be written as sizes by appending `B`, `K`, `M`, `G` and `T` to specify bytes, KB, MB, GB, and TB.
  - empty strings and `0` evaluate to `false`


## Properties

The following file properties are available:

| name        | description                                                       |
|-------------|-------------------------------------------------------------------|
| name        | name of the file                                                  |
| path        | full path of the file                                             |
| container   | path of the container (if inside an archive)                      |
| size        | file size (uncompressed)                                          |
| date        | modified date in YYYY-MM-DD format                                |
| time        | modified time in HH-MM-SS format                                  |
| ext         | short file extension (e.g., `txt`)                                |
| ext2        | long file extension (two parts, e.g., `tar.gz`)                   |
| type        | `file`, `dir`, or `link`                                          |
| archive     | archive type: `tar`, `zip`, `7z`, `rar` or empty                  |

Helper properties

| name        | description                                                       |
|-------------|-------------------------------------------------------------------|
| today       | today's date                                                      |
| mo          | last monday's date                                                |
| tu          | last tuesday's date                                               |
| we          | last wednesday's date                                             |
| th          | last thursday's date                                              |
| fr          | last friday's date                                                |
| sa          | last saturday's date                                              |
| su          | last sunday's date                                                |


## Supported archives

| name        | extensions                                                        |
|-------------|-------------------------------------------------------------------|
| tar         | `.tar`, `.tar.gz`, `.tgz`, `.tar.bz2`, `.tbz2`, `.tar.xz`, `.txz` |
| zip         | `.zip`                                                            |
| 7zip        | `.7z`                                                             |
| rar         | `.rar`                                                            |

> Note: use the flag -n (or --no-archive) to disable archive support. You can also use `'not archive'` in your query but this still requires zfind to open the archive.


## Actions

zfind does not implement actions like `find`, instead use `xargs -0` to execute commands:

```shell
zfind --no-archive 'name like "%.txt"' -0 | xargs -0 -L1 echo
```

zfind can also produce `--csv` (or `--csv-no-head`) that can be piped to other commands.


## Configuration

Set the environment variable `NO_COLOR` to disable color output.


## Installation


### Binary releases

You can download the official zfind binaries from the releases page and place it in your `PATH`.

- https://github.com/laktak/zfind/releases

### Homebrew (macOS and Linux)

For macOS and Linux it can also be installed via [Homebrew](https://formulae.brew.sh/formula/zfind):

```shell
brew install zfind
```

### Arch Linux

zfind is available in the AUR as [zfind](https://aur.archlinux.org/packages/zfind/):

```shell
paru -S zfind
```

### Build from Source

Building from the source requires Go.

- Either install it directly

```shell
go install github.com/laktak/zfind@latest
```

- or clone and build

```shell
git clone https://github.com/laktak/zfind
zfind/scripts/build
# output is here:
zfind/zfind
```


## zfind as a Go module

zfind is can also be used in other Go programs.

```
go get github.com/laktak/zfind
```

The library consists of two main packages:

- [filter](https://pkg.go.dev/github.com/laktak/zfind/filter): provides functionality for parsing and evaluating SQL-where filter expressions
- [find](https://pkg.go.dev/github.com/laktak/zfind/find): implements searching for files and directories.

For more information see the linked documentation on pkg.go.dev.

