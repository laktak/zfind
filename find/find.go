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
	"sort"
	"strings"
	"time"

	"github.com/bodgit/sevenzip"
	"github.com/laktak/zfind/filter"
	"github.com/nwaples/rardecode"
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

type FindError struct {
	Path string
	Err  error
}

func (e *FindError) Error() string { return e.Path + ": " + e.Err.Error() }

const (
	fieldName      = "name"
	fieldPath      = "path"
	fieldContainer = "container"
	fieldSize      = "size"
	fieldDate      = "date"
	fieldTime      = "time"
	fieldExt       = "ext"
	fieldExt2      = "ext2"
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
	fieldExt,
	fieldExt2,
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
		case fieldExt:
			return filter.TextValue(strings.TrimPrefix(filepath.Ext(file.Name), "."))
		case fieldExt2:
			return filter.TextValue(strings.TrimPrefix(ext2(file.Name), "."))
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

func listFilesInTar(fullpath string) ([]FileInfo, error) {
	f, err := os.Open(fullpath)
	if err != nil {
		return nil, &FindError{Path: fullpath, Err: err}
	}
	defer f.Close()

	var fr io.Reader = f
	switch {
	case strings.HasSuffix(fullpath, ".gz") || strings.HasSuffix(fullpath, ".tgz"):
		if fr, err = gzip.NewReader(f); err != nil {
			return nil, &FindError{Path: fullpath, Err: err}
		}
	case strings.HasSuffix(fullpath, ".bz2") || strings.HasSuffix(fullpath, ".tbz2"):
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
			return nil, &FindError{Path: fullpath, Err: err}
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

func getZipNameAndType(path string) (string, string) {
	if strings.HasSuffix(path, "/") {
		return path[:len(path)-1], "dir"
	} else {
		return path, "file"
	}
}

func listFilesInZip(fullpath string) ([]FileInfo, error) {
	f, err := os.Open(fullpath)
	if err != nil {
		return nil, &FindError{Path: fullpath, Err: err}
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, &FindError{Path: fullpath, Err: err}
	}
	zr, err := zip.NewReader(f, fi.Size())
	if err != nil {
		return nil, &FindError{Path: fullpath, Err: err}
	}

	var files []FileInfo
	for _, zf := range zr.File {
		rc, err := zf.Open()
		if err != nil {
			return nil, &FindError{Path: fullpath, Err: err}
		}
		defer rc.Close()
		name, t := getZipNameAndType(zf.Name)
		files = append(files, FileInfo{
			Container: fullpath,
			Path:      name,
			Name:      filepath.Base(name),
			Size:      int64(zf.UncompressedSize),
			ModTime:   zf.Modified,
			Type:      t,
			Archive:   "zip"})
	}
	return files, nil
}

func listFilesIn7Zip(fullpath string) ([]FileInfo, error) {

	r, err := sevenzip.OpenReader(fullpath)
	if err != nil {
		return nil, &FindError{Path: fullpath, Err: err}
	}
	defer r.Close()

	var files []FileInfo
	for _, h := range r.File {

		name, t := getZipNameAndType(h.Name)
		files = append(files, FileInfo{
			Container: fullpath,
			Path:      name,
			Name:      filepath.Base(name),
			Size:      h.FileInfo().Size(),
			ModTime:   h.Modified,
			Type:      t,
			Archive:   "7z"})
	}

	return files, nil
}

func listFilesInRar(fullpath string) ([]FileInfo, error) {

	r, err := rardecode.OpenReader(fullpath, "")
	if err != nil {
		return nil, &FindError{Path: fullpath, Err: err}
	}
	defer r.Close()

	var files []FileInfo
	for {
		h, err := r.Next()
		if err == io.EOF {
			break
		}

		t := "file"
		if h.IsDir {
			t = "dir"
		}

		files = append(files, FileInfo{
			Container: fullpath,
			Path:      h.Name,
			Name:      filepath.Base(h.Name),
			Size:      h.UnPackedSize,
			ModTime:   h.ModificationTime,
			Type:      t,
			Archive:   "rar"})
	}

	return files, nil
}

func findIn(param WalkParams, fi FileInfo) {

	fullpath := fi.Path

	if ok, err := param.Filter.Test(fi.Context()); err != nil {
		param.SendErr(&FindError{Path: fullpath, Err: err})
		return
	} else if ok {
		param.Chan <- fi
	}

	var files []FileInfo
	var err error = nil

	if fi.IsDir() || param.NoArchive {
		return
	}

	if strings.HasSuffix(fullpath, ".tar") ||
		strings.HasSuffix(fullpath, ".tar.gz") || strings.HasSuffix(fullpath, ".tgz") ||
		strings.HasSuffix(fullpath, ".tar.bz2") || strings.HasSuffix(fullpath, ".tbz2") {
		files, err = listFilesInTar(fullpath)
	} else if strings.HasSuffix(fullpath, ".zip") {
		files, err = listFilesInZip(fullpath)
	} else if strings.HasSuffix(fullpath, ".7z") {
		files, err = listFilesIn7Zip(fullpath)
	} else if strings.HasSuffix(fullpath, ".rar") {
		files, err = listFilesInRar(fullpath)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	if err != nil {
		param.SendErr(err)
	} else {
		for _, fi2 := range files {
			if ok, err := param.Filter.Test(fi2.Context()); err != nil {
				param.SendErr(&FindError{Path: fullpath, Err: err})
				return
			} else if ok {
				param.Chan <- fi2
			}
		}
	}
}

type WalkParams struct {
	Chan           chan FileInfo
	Err            chan string
	Filter         *filter.FilterExpression
	FollowSymlinks bool
	NoArchive      bool
}

func (wp WalkParams) SendErr(err error) {
	serr := fmt.Sprintf("%v", err)
	wp.Err <- serr
}

func Walk(root string, param WalkParams) {
	fsWalk(root, param.FollowSymlinks, func(fi *FileInfo, err error) {
		if err == nil {
			findIn(param, *fi)
		} else {
			param.SendErr(err)
		}
	})
}
