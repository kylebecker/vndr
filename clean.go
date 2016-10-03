package main

import (
	"go/build"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LK4D4/vndr/godl"
)

func isCDir(path string) bool {
	fis, err := ioutil.ReadDir(path)
	if err != nil {
		return false
	}
	var hFound bool
	for _, fi := range fis {
		ext := filepath.Ext(fi.Name())
		if ext == ".cc" || ext == ".cpp" || ext == ".py" {
			return false
		}
		if ext == ".h" {
			hFound = true
		}
	}
	return hFound
}

func isGoFile(path string) bool {
	ext := filepath.Ext(path)
	return ext == ".go" || ext == ".c" || ext == ".h" || ext == ".s" || ext == ".proto"
}

var licenseFiles = map[string]bool{
	"LICENSE": true,
	"COPYING": true,
	"PATENTS": true,
	"NOTICE":  true,
}

// cleanVendor removes files from unused pacakges and non-go files
func cleanVendor(vendorDir string, realDeps []*build.Package) error {
	realPaths := make(map[string]bool)
	for _, pkg := range realDeps {
		realPaths[pkg.Dir] = true
	}
	var paths []string
	err := filepath.Walk(vendorDir, func(path string, i os.FileInfo, err error) error {
		if path == vendorDir {
			return nil
		}
		if err != nil {
			return nil
		}
		if strings.HasPrefix(i.Name(), ".") || strings.HasPrefix(i.Name(), "_") {
			return os.RemoveAll(path)
		}
		if i.IsDir() {
			if i.Name() == "testdata" {
				return os.RemoveAll(path)
			}
			if isCDir(path) {
				realPaths[path] = true
				return nil
			}
			if !realPaths[path] {
				paths = append(paths, path)
			}
			return nil
		}

		// keep files for licenses
		if licenseFiles[i.Name()] {
			return nil
		}
		// remove files from non-deps, non-go files and test files
		if !realPaths[filepath.Dir(path)] || !isGoFile(path) || strings.HasSuffix(path, "_test.go") {
			return os.Remove(path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	sort.Sort(sort.Reverse(sort.StringSlice(paths)))
	// iterate over paths (longest first)
	for _, p := range paths {
		// at this point we cleaned all files from unused deps dirs
		lst, err := ioutil.ReadDir(p)
		if err != nil {
			return err
		}
		if len(lst) != 0 {
			continue
		}
		// remove all directories if they're not in dependency paths
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}
	return nil
}

func cleanVCS(v *godl.VCS) error {
	if err := os.RemoveAll(filepath.Join(v.Root, "."+v.Type)); err != nil {
		return err
	}
	return filepath.Walk(v.Root, func(path string, i os.FileInfo, err error) error {
		if path == vendorDir {
			return nil
		}
		if !i.IsDir() {
			return nil
		}
		name := i.Name()
		if name == "vendor" || name == "Godeps" || name == "_vendor" {
			if err := os.RemoveAll(path); err != nil {
				return err
			}
			return filepath.SkipDir
		}
		return nil
	})
}
