package find

import (
	"os"
	"path/filepath"
	"sort"
)

type WalkFunc func(file *FileInfo, err error)

type WalkError struct {
	Path string
	Err  error
}

func (e *WalkError) Error() string { return e.Path + ": " + e.Err.Error() }

func fsWalk(root string, followSymlinks bool, report WalkFunc) {
	walk(root, root, followSymlinks, report)
}

func makeFileInfo(fullpath string, file os.FileInfo) FileInfo {
	ft := file.Mode().Type()
	t := "file"
	if ft&os.ModeDir != 0 {
		t = "dir"
	} else if ft&os.ModeSymlink != 0 {
		t = "link"
	}

	return FileInfo{
		Path:    fullpath,
		Name:    file.Name(),
		Size:    file.Size(),
		ModTime: file.ModTime(),
		Type:    t,
	}
}

func readDirNames(dirname string) ([]string, error) {
	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	names, err := f.Readdirnames(-1)
	f.Close()
	if err != nil {
		return nil, err
	}
	sort.Strings(names)
	return names, nil
}

func walk(path string, virtPath string, followSymlinks bool, report WalkFunc) {

	osFileInfo, err := os.Lstat(path)
	if err != nil {
		report(nil, &WalkError{Path: path, Err: err})
	} else {
		fi := makeFileInfo(virtPath, osFileInfo)
		if fi.IsDir() {
			report(&fi, nil)
		} else if fi.Type == "link" && followSymlinks {
			rpath, err := filepath.EvalSymlinks(path)
			if err != nil {
				report(nil, &WalkError{Path: path, Err: err})
				return
			}
			path = rpath
			fi.Type = "dir"
			report(&fi, nil)
		} else {
			// file
			report(&fi, nil)
			return
		}

		names, err := readDirNames(path)
		if err != nil {
			report(nil, &WalkError{Path: path, Err: err})
		} else {
			for _, name := range names {
				rfilename := filepath.Join(path, name)
				filename := filepath.Join(virtPath, name)

				walk(rfilename, filename, followSymlinks, report)
			}
		}
	}
}

func ext2(path string) string {
	for i, n := len(path)-1, 0; i >= 0 && !os.IsPathSeparator(path[i]); i-- {
		if path[i] == '.' {
			n += 1
			if n == 2 {
				return path[i:]
			}
		}
	}
	return ""
}
