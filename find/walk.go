package find

import (
	"os"
	"path/filepath"
	"sort"
)

type WalkFunc func(file *FileInfo, err error)

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
		report(nil, err)
	} else {
		fi := makeFileInfo(virtPath, osFileInfo)
		if fi.IsDir() {
			report(&fi, nil)
			//err = walk(path, virtPath, followSymlinks, report)
		} else if fi.Type == "link" && followSymlinks {
			path, err = filepath.EvalSymlinks(path)
			if err != nil {
				report(nil, err)
				return
			}
			fi.Type = "dir"
			report(&fi, nil)
			//err = walk(path, virtPath, followSymlinks, report)
		} else {
			// file
			report(&fi, nil)
			return
		}

		names, err := readDirNames(path)
		if err != nil {
			report(nil, err)
		} else {
			for _, name := range names {
				rfilename := filepath.Join(path, name)
				filename := filepath.Join(virtPath, name)

				walk(rfilename, filename, followSymlinks, report)
			}
		}
	}
}
