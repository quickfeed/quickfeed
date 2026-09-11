package ci

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// copyDir recursively copies the directory tree rooted at src into dst,
// creating dst and its parents if necessary. File permission bits are
// preserved; directories additionally get the owner's read, write, and execute
// bits so that the copy can proceed into them.
//
// Any .git directory is skipped: only the submitted code is tested, and the
// repository metadata is both large and irrelevant to the test run. Irregular
// files, such as symbolic links and sockets, are skipped as well; a test run
// must not follow a link out of the copied tree.
func copyDir(src, dst string) error {
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

// copyFile copies the regular file src to dst, creating or truncating dst with
// the given permission bits.
func copyFile(src, dst string, mode fs.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
