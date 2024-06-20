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
	"github.com/ulikunitz/xz"
)

// FileInfo is a type that represents information about a file or directory.
type FileInfo struct {
	Name      string
	Path      string
	ModTime   time.Time
	Size      int64
	Type      string
	Container string
	Archive   string
}

// IsDir returns a boolean value indicating if the FileInfo instance is a
// directory.
func (fi FileInfo) IsDir() bool { return fi.Type == "dir" }

func (fi FileInfo) fromSymlink(fi2 FileInfo) FileInfo {
	return FileInfo{
		Name:    fi.Name,
		Path:    fi.Path,
		ModTime: fi2.ModTime,
		Size:    fi2.Size,
		Type:    fi2.Type,
	}
}

// FindError is a type that represents an error that occurred during a file search.
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

// Fields is a slice of the constants that address fields in the FileInfo type.
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

// Context is a method of the FileInfo type that returns a VariableGetter function
// that can be used to retrieve the values of the fields of the file or directory
// represented by the FileInfo instance.
//
// It also generates helper properties like "today".
func (file FileInfo) Context() filter.VariableGetter {
	return func(name string) *filter.Value {
		switch strings.ToLower(name) {
		case fieldName:
			return filter.TextValue(file.Name)
		case fieldPath:
			return filter.TextValue(file.Path)
		case fieldDate:
			return filter.TextValue(file.ModTime.Format(time.DateOnly))
		case fieldTime:
			return filter.TextValue(file.ModTime.Format(time.TimeOnly))
		case fieldSize:
			return filter.NumberValue(file.Size)
		case fieldExt:
			return filter.TextValue(strings.TrimPrefix(filepath.Ext(file.Name), "."))
		case fieldExt2:
			return filter.TextValue(strings.TrimPrefix(ext2(file.Name), "."))
		case fieldType:
			return filter.TextValue(file.Type)
		case fieldContainer:
			return filter.TextValue(file.Container)
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
	case strings.HasSuffix(fullpath, ".xz") || strings.HasSuffix(fullpath, ".txz"):
		if fr, err = xz.NewReader(f); err != nil {
			return nil, &FindError{Path: fullpath, Err: err}
		}
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
				Name:      filepath.Base(h.Name),
				Path:      h.Name,
				ModTime:   h.ModTime,
				Size:      h.Size,
				Type:      t,
				Container: fullpath,
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
			Name:      filepath.Base(name),
			Path:      name,
			ModTime:   zf.Modified,
			Size:      int64(zf.UncompressedSize),
			Type:      t,
			Container: fullpath,
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
			Name:      filepath.Base(name),
			Path:      name,
			ModTime:   h.Modified,
			Size:      h.FileInfo().Size(),
			Type:      t,
			Container: fullpath,
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
			Name:      filepath.Base(h.Name),
			Path:      h.Name,
			ModTime:   h.ModificationTime,
			Size:      h.UnPackedSize,
			Type:      t,
			Container: fullpath,
			Archive:   "rar"})
	}

	return files, nil
}

func findIn(param WalkParams, fi FileInfo) {

	fullpath := fi.Path

	if ok, err := param.Filter.Test(fi.Context()); err != nil {
		param.sendErr(&FindError{Path: fullpath, Err: err})
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
		strings.HasSuffix(fullpath, ".tar.bz2") || strings.HasSuffix(fullpath, ".tbz2") ||
		strings.HasSuffix(fullpath, ".tar.xz") || strings.HasSuffix(fullpath, ".txz") {
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
		param.sendErr(err)
	} else {
		for _, fi2 := range files {
			if ok, err := param.Filter.Test(fi2.Context()); err != nil {
				param.sendErr(&FindError{Path: fullpath, Err: err})
				return
			} else if ok {
				param.Chan <- fi2
			}
		}
	}
}

// WalkParams is used to specify the parameters for a file search.
type WalkParams struct {
	// Chan is the channel that is used to send the results of the search.
	Chan chan FileInfo
	// Err is the channel that is used to send error messages.
	Err chan string
	// Filter is the filter expression that is used to filter the results of the search.
	Filter *filter.FilterExpression
	// FollowSymlinks specifies whether symbolic links should be followed during the search.
	FollowSymlinks bool
	// NoArchive specifies whether archives should be skipped during the search.
	NoArchive bool
}

// Sends an error message to the error channel of the WalkParams instance.
func (wp WalkParams) sendErr(err error) {
	serr := fmt.Sprintf("%v", err)
	wp.Err <- serr
}

// Walk is a function that performs a file search starting at the given root
// directory. See WalkParams to control the behavior of the search.
func Walk(root string, param WalkParams) {
	fsWalk(root, param.FollowSymlinks, func(fi *FileInfo, err error) {
		if err == nil {
			findIn(param, *fi)
		} else {
			param.sendErr(err)
		}
	})
}
