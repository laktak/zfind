
# zfind

zfind allows you to search for files, including tar/zip, using a SQL-WHERE clause filter.

The tool is already useable although it is missing some features (and a README).

```
$ zfind -h
Usage: zfind [<path> ...] [flags]

Arguments:
  [<path> ...]    Paths to search.

Flags:
  -h, --help                      Show context-sensitive help.
  -H, --filter-help               Show where-filter help.
  -w, --where=STRING              The where-filter (using sql-where syntax, see -H).
  -l, --long                      Show long listing.
      --csv                       Show listing as csv.
      --archive-separator="//"    Separator between the archive name and the file inside
  -L, --follow-symlinks           Follow symbolic links.
  -V, --version                   Show version.
```

```
$ zfind -H

zfind uses a filter syntax that closely resembles an SQL-WHERE clause.

eg.
  name="foo.txt"
  name like "bar%"
  name like "%.txt" and archive="tar"
  name in ("foo", "bar") and type="file"
  date between "2000-01-01" and "2010-12-31"

The following 'columns' are available:

  name        name of the file
  path        full path of the file
              (relative to the file system or archive)
  size        file size (uncompressed)
  date        modified date in YYYY-MM-DD format
  time        modified time in HH-MM-SS format
  type        file|dir|link
  archive     tar|zip if inside a container
  container   path of container (if any)

```

