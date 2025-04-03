package walkentry

import (
	"io/fs"
	"iter"
	"path"
	"strings"
)

// WalkEntry represents a file system entry encountered during traversal.
type WalkEntry struct {
	Path      string // path relative to root
	Name      string // name without extension
	Ext       string // file extension (with dot)
	IsDir     bool   // whether entry is a directory
	IsRegular bool   // whether entry is a regular file
}

// isFile returns true if the entry is a regular file
func (e *WalkEntry) isFile() bool {
	return e.IsRegular
}

// isFileWithExt returns true if the entry is a file with the specified extension
func (e *WalkEntry) isFileWithExt(ext string) bool {
	return e.isFile() && e.Ext == ext
}

// isJSONFile returns true if the entry is a JSON file
func (e *WalkEntry) IsJSONFile() bool {
	return e.isFileWithExt(".json")
}

// newWalkEntry creates a new WalkEntry from a directory entry and root path
func newWalkEntry(entry fs.DirEntry, root string) WalkEntry {
	name := entry.Name()
	ext := path.Ext(name)

	return WalkEntry{
		Path:      cleanPath(path.Join(root, name)),
		Name:      strings.TrimSuffix(name, ext),
		Ext:       ext,
		IsDir:     entry.IsDir(),
		IsRegular: entry.Type().IsRegular(),
	}
}

// cleanPath removes the "./" prefix from the path if present
func cleanPath(p string) string {
	return strings.TrimPrefix(p, "./")
}

// directoryWalker handles walking through a directory
type directoryWalker struct {
	fsys fs.FS
	root string
}

// newDirectoryWalker creates a new directoryWalker instance
func newDirectoryWalker(fsys fs.FS, root string) directoryWalker {
	return directoryWalker{
		fsys: fsys,
		root: root,
	}
}

// walk traverses the directory tree and yields entries
func (w *directoryWalker) walk(yield func(WalkEntry, error) bool) bool {
	// Read directory entries
	entries, err := fs.ReadDir(w.fsys, w.root)
	if err != nil {
		return yield(WalkEntry{}, err)
	}

	// Process each entry
	for _, entry := range entries {
		if !w.processEntry(entry, yield) {
			return false
		}
	}

	return true
}

// processEntry handles a single directory entry
func (w *directoryWalker) processEntry(entry fs.DirEntry, yield func(WalkEntry, error) bool) bool {
	// Create and yield the current entry
	walkEntry := newWalkEntry(entry, w.root)
	if !yield(walkEntry, nil) {
		return false
	}

	// Handle subdirectory recursion
	if entry.IsDir() {
		subPath := path.Join(w.root, entry.Name())
		return w.walkSubdirectory(subPath, yield)
	}

	return true
}

// walkSubdirectory handles recursive directory traversal
func (w *directoryWalker) walkSubdirectory(subPath string, yield func(WalkEntry, error) bool) bool {
	subWalker := newDirectoryWalker(w.fsys, subPath)
	return subWalker.walk(yield)
}

// walkFS returns an iterator that walks through the filesystem starting from root
func WalkFS(fsys fs.FS, root string) iter.Seq2[WalkEntry, error] {
	return func(yield func(WalkEntry, error) bool) {
		walker := newDirectoryWalker(fsys, root)
		walker.walk(yield)
	}
}
