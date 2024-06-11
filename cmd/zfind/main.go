package main

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/fatih/color"
	"github.com/laktak/zfind/filter"
	"github.com/laktak/zfind/find"
)

var appVersion = "vdev"

func printFiles(ch chan find.FileInfo, long bool, archSep string, lineSep []byte) {
	for file := range ch {
		name := ""
		if file.Container != "" {
			name = file.Container + archSep
		}
		name += file.Path
		if long {
			size := filter.FormatSize(file.Size)
			fmt.Fprintf(os.Stdout, "%s %10s %s", file.ModTime.Format("2006-01-02 15:04:05"), size, name)
		} else {
			fmt.Fprint(os.Stdout, name)
		}
		os.Stdout.Write(lineSep)
	}
}

func printCsv(ch chan find.FileInfo) error {
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
		Long             bool     `short:"l" help:"Show long listing format."`
		Csv              bool     `help:"Show listing as csv."`
		ArchiveSeparator string   `help:"Separator between the archive name and the file inside" default:"//"`
		FollowSymlinks   bool     `short:"L" help:"Follow symbolic links."`
		NoArchive        bool     `short:"n" help:"Disables archive support."`
		Print0           bool     `name:"print0" short:"0" help:"Use a null character instead of the newline character, to be used with the -0 option of xargs."`
		Version          bool     `short:"V" help:"Show version."`
		XWhere           string   `name:"where" short:"w" help:"(removed) this option has moved to the <where> argument"`
		Where            string   `arg:"" name:"where" optional:"" help:"The filter using sql-where syntax (see -H). Use '-' to skip when providing a path."`
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

	if cli.XWhere != "" {
		fmt.Println("error: the -w option has been replaced by the <where> argument. Usage: zfind <where> [<path>]")
		os.Exit(1)
	}

	if cli.Where == "" || cli.Where == "-" {
		cli.Where = "1"
	}

	lineSep := []byte("\n")
	if cli.Print0 {
		lineSep = []byte{0}
	}

	if len(cli.Paths) == 0 {
		cli.Paths = []string{"."}
	}

	filter, err := filter.CreateFilter(cli.Where)
	arg.FatalIfErrorf(err)

	done := make(chan bool)
	ch := make(chan find.FileInfo)
	errChan := make(chan string)

	// start search
	go func() {
		for _, searchPath := range cli.Paths {
			find.Walk(searchPath, find.WalkParams{
				Chan:           ch,
				Err:            errChan,
				Filter:         filter,
				FollowSymlinks: cli.FollowSymlinks,
				NoArchive:      cli.NoArchive})
		}
		close(ch)
		close(errChan)
	}()

	// print results
	go func() {
		if cli.Csv {
			arg.FatalIfErrorf(printCsv(ch))
		} else {
			printFiles(ch, cli.Long, cli.ArchiveSeparator, lineSep)
		}
		done <- true
	}()

	// print errors
	hasErr := false
	var errCol = color.New(color.FgRed).SprintFunc()
	for errmsg := range errChan {
		fmt.Fprintln(color.Error, errCol("error: "+errmsg))
		hasErr = true
	}

	// wait for output to finish
	<-done

	if hasErr {
		fmt.Fprintln(color.Error, errCol("errors were encountered!"))
		os.Exit(1)
	}
}
