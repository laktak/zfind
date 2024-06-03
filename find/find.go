package find

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
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

const (
	fieldContainer = "container"
	fieldName      = "name"
	fieldPath      = "path"
	fieldSize      = "size"
	fieldDate      = "date"
	fieldTime      = "time"
	fieldType      = "type"
	fieldArchive   = "archive"
)

var Fields = [...]string{
	fieldContainer,
	fieldName,
	fieldPath,
	fieldSize,
	fieldDate,
	fieldTime,
	fieldType,
	fieldArchive,
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
			return filter.TextValue(file.ModTime.Format("2006-01-02 15:04:05"))
		case fieldTime:
			return filter.TextValue(file.ModTime.Format("15:04:05"))
		case fieldType:
			return filter.TextValue(file.Type)
		case fieldArchive:
			return filter.TextValue(file.Archive)
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

func findIn(filter *filter.FilterExpression, fullpath string, file os.FileInfo, found chan FileInfo, err error) error {
	if err != nil {
		return err
	}

	filename := file.Name()
	isTar, isZip := false, false

	if !file.IsDir() {
		isTar = strings.HasSuffix(filename, ".tar") || strings.HasSuffix(filename, ".tar.gz") || strings.HasSuffix(filename, ".tgz") || strings.HasSuffix(filename, ".tar.bz2") || strings.HasSuffix(filename, ".tbz2")
		isZip = strings.HasSuffix(filename, ".zip")
	}

	switch {
	case isTar:
		files, err := listFilesInTar(fullpath, filename)
		if err != nil {
			return err
		}
		for _, file2 := range files {
			if ok, err := filter.Test(file2.Context()); err != nil {
				return err
			} else if ok {
				found <- file2
			}
		}

	case isZip:
		files, err := listFilesInZip(fullpath, filename)
		if err != nil {
			return err
		}
		for _, file2 := range files {
			if ok, err := filter.Test(file2.Context()); err != nil {
				return err
			} else if ok {
				found <- file2
			}
		}

	default:
		ft := file.Mode().Type()
		t := "file"
		if ft&os.ModeDir != 0 {
			t = "dir"
		} else if ft&os.ModeSymlink != 0 {
			t = "link"
		}

		file2 := FileInfo{
			Path:    fullpath,
			Name:    file.Name(),
			Size:    file.Size(),
			ModTime: file.ModTime(),
			Type:    t,
		}

		if ok, err := filter.Test(file2.Context()); err != nil {
			return err
		} else if ok {
			found <- file2
		}
	}
	return nil
}

func Walk(filter *filter.FilterExpression, root string, found chan FileInfo) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		return findIn(filter, path, info, found, err)
	})
}
