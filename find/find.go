package find

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/laktak/zfind/filter"
)

type FileInfo struct {
	Container string
	Name      string
	Path      string
	Size      int64
	Type      string
	Archive   string
	ModTime   time.Time
}

func (fi FileInfo) IsDir() bool { return fi.Type == "dir" }

const (
	fieldName      = "name"
	fieldPath      = "path"
	fieldContainer = "container"
	fieldSize      = "size"
	fieldDate      = "date"
	fieldTime      = "time"
	fieldType      = "type"
	fieldArchive   = "archive"
	fieldToday     = "today"
)

var Fields = [...]string{
	fieldName,
	fieldPath,
	fieldContainer,
	fieldSize,
	fieldDate,
	fieldTime,
	fieldType,
	fieldArchive,
	// not exported
	// fieldToday
}

func (file FileInfo) Context() filter.VariableGetter {
	return func(name string) *filter.Value {
		switch strings.ToLower(name) {
		case fieldContainer:
			return filter.TextValue(file.Container)
		case fieldName:
			return filter.TextValue(file.Name)
		case fieldPath:
			return filter.TextValue(file.Path)
		case fieldSize:
			return filter.NumberValue(file.Size)
		case fieldDate:
			return filter.TextValue(file.ModTime.Format(time.DateOnly))
		case fieldTime:
			return filter.TextValue(file.ModTime.Format(time.TimeOnly))
		case fieldType:
			return filter.TextValue(file.Type)
		case fieldArchive:
			return filter.TextValue(file.Archive)
		case fieldToday:
			return filter.TextValue(time.Now().Format(time.DateOnly))
		default:
			return nil
		}
	}
}

func listFilesInZip(fullpath string, filename string) ([]FileInfo, error) {
	f, err := os.Open(fullpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		return nil, err
	}

	var files []FileInfo
	for _, zf := range zr.File {
		rc, err := zf.Open()
		if err != nil {
			return nil, err
		}
		defer rc.Close()
		files = append(files, FileInfo{
			Container: fullpath,
			Path:      zf.Name,
			Name:      filepath.Base(zf.Name),
			Size:      int64(zf.UncompressedSize),
			ModTime:   zf.Modified,
			Type:      "file",
			Archive:   "zip"})
	}
	return files, nil
}

func listFilesInTar(fullpath string, filename string) ([]FileInfo, error) {
	f, err := os.Open(fullpath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var fr io.Reader = f
	switch {
	case strings.HasSuffix(filename, ".gz") || strings.HasSuffix(filename, ".tgz"):
		if fr, err = gzip.NewReader(f); err != nil {
			return nil, err
		}
	case strings.HasSuffix(filename, ".bz2") || strings.HasSuffix(filename, ".tbz2"):
		fr = bzip2.NewReader(f)
	}

	r := tar.NewReader(fr)

	var files []FileInfo
	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		switch h.Typeflag {
		case tar.TypeReg, tar.TypeDir, tar.TypeSymlink:
			t := "file"
			if h.Typeflag == tar.TypeDir {
				t = "dir"
			} else if h.Typeflag == tar.TypeSymlink {
				t = "link"
			}

			files = append(files, FileInfo{
				Container: fullpath,
				Path:      h.Name,
				Name:      filepath.Base(h.Name),
				Size:      h.Size,
				ModTime:   h.ModTime,
				Type:      t,
				Archive:   "tar"})
		}
	}

	return files, nil
}

func findIn(param WalkParams, fi FileInfo) {

	if ok, err := param.Filter.Test(fi.Context()); err != nil {
		param.SendErr(err)
		return
	} else if ok {
		param.Chan <- fi
	}

	filename := fi.Name
	isTar, isZip := false, false
	if !fi.IsDir() {
		isTar = strings.HasSuffix(filename, ".tar") || strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") || strings.HasSuffix(filename, ".tar.bz2") || strings.HasSuffix(filename, ".tbz2")
		isZip = strings.HasSuffix(filename, ".zip")
	}

	var files []FileInfo
	var err error = nil

	switch {
	case isTar:
		files, err = listFilesInTar(fi.Path, filename)

	case isZip:
		files, err = listFilesInZip(fi.Path, filename)
	default:
		return
	}

	if err != nil {
		param.SendErr(err)
		return
	}
	for _, fi2 := range files {
		if ok, err := param.Filter.Test(fi2.Context()); err != nil {
			param.SendErr(err)
			return
		} else if ok {
			param.Chan <- fi2
		}
	}
}

type WalkParams struct {
	Chan           chan FileInfo
	Err            chan string
	Filter         *filter.FilterExpression
	FollowSymlinks bool
}

func (wp WalkParams) SendErr(err error) {
	serr := fmt.Sprintf("%v", err)
	wp.Err <- serr
}

func Walk(root string, param WalkParams) error {
	return fsWalk(root, param.FollowSymlinks, func(fi *FileInfo, err error) {
		if err == nil {
			findIn(param, *fi)
		} else {
			param.SendErr(err)
		}
	})
}
