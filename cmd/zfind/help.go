package main

var filter_help = `
zfind uses a filter syntax that is very similar to an SQL-WHERE clause.

Examples:

	# find files smaller than 10KB
	zfind 'size<10k'

	# find files modified before 2010 inside a tar
	zfind 'date<"2010" and archive="tar"'

	# find files named *.go and modified today
	zfind 'name like "%.go" and date=today'

	# find directories named foo and bar
	zfind 'name in ("foo", "bar") and type="dir"'

	# search for all README.md files and show in long listing format
	zfind 'name="README.md"' -l

	# show results in csv format
	zfind --csv

The following file properies are available:

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

For more details go to https://github.com/laktak/zfind
`
