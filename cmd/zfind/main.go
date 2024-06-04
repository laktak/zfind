package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/laktak/zfind/filter"
	"github.com/laktak/zfind/find"
)

var appVersion = "vdev"

func PrintFiles(ch chan find.FileInfo, long bool, archSep string) {
	for file := range ch {
		name := ""
		if file.Container != "" {
			name = file.Container + archSep
		}
		name += file.Path
		if long {
			size := filter.FormatSize(file.Size)
			fmt.Printf("%s %10s %s\n", file.ModTime.Format("2006-01-02 15:04:05"), size, name)
		} else {
			fmt.Println(name)
		}
	}
}

func PrintCsv(ch chan find.FileInfo) error {
	writer := csv.NewWriter(os.Stdout)

	if err := writer.Write(find.Fields[:]); err != nil {
		return err
	}

	for file := range ch {
		var record []string
		getter := file.Context()
		for _, field := range find.Fields {
			value := getter(field)
			record = append(record, (*value).String())
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	writer.Flush()

	if err := writer.Error(); err != nil {
		return err
	}

	return nil
}

func main() {
	var cli struct {
		FilterHelp       bool     `short:"H" help:"Show where-filter help."`
		Where            string   `short:"w" help:"The where-filter (using sql-where syntax, see -H)."`
		Long             bool     `short:"l" help:"Show long listing."`
		Csv              bool     `help:"Show listing as csv."`
		ArchiveSeparator string   `help:"Separator between the archive name and the file inside" default:"//"`
		FollowSymlinks   bool     `short:"L" help:"Follow symbolic links."`
		Version          bool     `short:"V" help:"Show version."`
		Paths            []string `arg:"" name:"path" optional:"" help:"Paths to search."`
	}

	arg := kong.Parse(&cli)

	if cli.FilterHelp {
		fmt.Println(filter_help)
		os.Exit(0)
	}

	if cli.Version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	if cli.Where == "" {
		cli.Where = "1"
	}

	if len(cli.Paths) == 0 {
		cli.Paths = []string{"."}
	}

	filter, err := filter.CreateFilter(cli.Where)
	arg.FatalIfErrorf(err)

	ch := make(chan find.FileInfo)
	errChan := make(chan string)

	go func() {
		for _, searchPath := range cli.Paths {
			err = find.Walk(searchPath, find.WalkParams{
				Chan:           ch,
				Err:            errChan,
				Filter:         filter,
				FollowSymlinks: cli.FollowSymlinks})
			arg.FatalIfErrorf(err)
		}
		close(ch)
	}()

	go func() {
		for errmsg := range errChan {
			fmt.Fprintln(os.Stderr, errmsg)
		}
		close(errChan)
	}()

	if cli.Csv {
		arg.FatalIfErrorf(PrintCsv(ch))
	} else {
		PrintFiles(ch, cli.Long, cli.ArchiveSeparator)
	}
}
