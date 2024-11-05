package main

var headerHelp = `Search for files, including inside tar, zip, 7z and rar archives.
 zfind makes finding files easy with a filter syntax that is similar to an SQL-WHERE clause.
 For examples run "zfind -H" or go to
 https://github.com/laktak/zfind
`

var filterHelp = `
zfind uses a filter syntax that is very similar to an SQL-WHERE clause.

Examples:

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

The following file properties are available:

  name        name of the file
  path        full path of the file
  size        file size (uncompressed)
  date        modified date in YYYY-MM-DD format
  time        modified time in HH-MM-SS format
  ext         short file extension (e.g. 'txt')
  ext2        long file extension (two parts, e.g. 'tar.gz')
  type        file|dir|link
  archive     archive type tar|zip|7z|rar if inside a container
  container   path of container (if any)

Helper properties

  today       todays date
  mo          last monday's date
  tu          last tuesday's date
  we          last wednesday's date
  th          last thursday's date
  fr          last friday's date
  sa          last saturday's date
  su          last sunday's date

For more details go to https://github.com/laktak/zfind
`
