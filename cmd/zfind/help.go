package main

var filter_help = `
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


Helper columns:

  today       todays date (e.g. for "date=today")

`
