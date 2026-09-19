// Package fileop holds file system operations shared across QuickFeed.
package fileop

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CopyDir recursively copies the directory tree rooted at src to dst. dst and
// its missing parents are created; an existing dst is written into. File
// permission bits are preserved, and directories additionally get the owner's
// read, write, and execute bits so that the copy can proceed into them.
//
// Any .git directory is skipped, wherever it occurs in the tree: the copy is a
// fresh working tree, not a clone. Irregular files, such as symbolic links and
// sockets, are skipped as well, so that a copy of untrusted code cannot reach
// outside the copied tree through a link.
//
// The source must exist, and the two directories must not overlap; see
// resolveCopyPaths.
func CopyDir(src, dst string) error {
	src, dst, err := resolveCopyPaths(src, dst)
	if err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			// WalkDir was unable to read path; stop walking the tree.
			return err
		}
		if entry.IsDir() && entry.Name() == ".git" {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if entry.IsDir() {
			return os.MkdirAll(target, info.Mode().Perm()|0o700)
		}
		if !info.Mode().IsRegular() {
			return nil
		}
		return copyFile(path, target, info.Mode().Perm())
	})
}

// resolveCopyPaths returns src and dst as absolute paths with their symbolic
// links resolved, and fails if the two overlap. A destination inside the source
// would make CopyDir's walk descend into the copy it is writing, and a source
// inside the destination would be copied into itself.
//
// The source must exist and is resolved as a whole; the destination need not,
// and is resolved as far as it exists. Resolving is what makes the comparison
// meaningful on macOS, where both /tmp and $TMPDIR are symbolic links.
func resolveCopyPaths(src, dst string) (string, string, error) {
	src, err := filepath.Abs(src)
	if err != nil {
		return "", "", err
	}
	if src, err = filepath.EvalSymlinks(src); err != nil {
		return "", "", err
	}
	dst, err = filepath.Abs(dst)
	if err != nil {
		return "", "", err
	}
	if dst, err = evalExistingPrefix(dst); err != nil {
		return "", "", err
	}
	switch {
	case dst == src:
		return "", "", fmt.Errorf("destination directory %q is the source directory", dst)
	case strings.HasPrefix(dst, withSeparator(src)):
		return "", "", fmt.Errorf("destination directory %q is inside the source directory %q", dst, src)
	case strings.HasPrefix(src, withSeparator(dst)):
		return "", "", fmt.Errorf("source directory %q is inside the destination directory %q", src, dst)
	}
	return src, dst, nil
}

// evalExistingPrefix resolves the symbolic links in the longest existing
// prefix of the absolute path and rejoins the elements that do not exist yet,
// so that a destination CopyDir has still to create compares correctly against
// an existing source.
func evalExistingPrefix(path string) (string, error) {
	missing := ""
	for {
		resolved, err := filepath.EvalSymlinks(path)
		if err == nil {
			return filepath.Join(resolved, missing), nil
		}
		if !errors.Is(err, fs.ErrNotExist) {
			return "", err
		}
		parent := filepath.Dir(path)
		if parent == path {
			// Reached the root without finding an existing directory.
			return "", err
		}
		missing = filepath.Join(filepath.Base(path), missing)
		path = parent
	}
}

// withSeparator returns dir with a trailing separator, so that a prefix
// comparison against it cannot match a sibling: /tmp/a does not contain
// /tmp/ab. The root directory already ends in a separator.
func withSeparator(dir string) string {
	if strings.HasSuffix(dir, string(filepath.Separator)) {
		return dir
	}
	return dir + string(filepath.Separator)
}

// copyFile copies the regular file src to dst, creating or truncating dst with
// the given permission bits. A failure to close either file is reported, since
// a failed close of the destination can mean that the data never reached disk.
func copyFile(src, dst string, mode fs.FileMode) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := in.Close(); err == nil {
			err = closeErr
		}
	}()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(out, in)
	return err
}
