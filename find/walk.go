package find

import (
	"os"
	"path/filepath"
	"sort"
)

type WalkFunc func(file *FileInfo, err error)

func fsWalk(root string, followSymlinks bool, walkFn WalkFunc) error {
	osFileInfo, err := os.Lstat(root)
	if err != nil {
		walkFn(nil, err)
	} else {
		fi := makeFileInfo(root, osFileInfo)
		walkFn(&fi, nil)
	}
	return walk(root, root, followSymlinks, walkFn)
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

func walk(path string, virtPath string, followSymlinks bool, walkFn WalkFunc) error {

	names, err := readDirNames(path)
	if err != nil {
		walkFn(nil, err)
		return nil
	}

	for _, name := range names {
		rfilename := filepath.Join(path, name)
		filename := filepath.Join(virtPath, name)
		osFileInfo, err := os.Lstat(rfilename)
		if err != nil {
			walkFn(nil, err)
		} else {
			fi := makeFileInfo(filename, osFileInfo)
			if fi.IsDir() {
				walkFn(&fi, nil)
				err = walk(rfilename, filename, followSymlinks, walkFn)
			} else if fi.Type == "link" && followSymlinks {
				rfilename, err = filepath.EvalSymlinks(rfilename)
				fi.Type = "dir"
				walkFn(&fi, nil)
				err = walk(rfilename, filename, followSymlinks, walkFn)
			} else {
				walkFn(&fi, nil)
			}
			if err != nil {
				walkFn(nil, err)
			}
		}
	}
	return nil
}
