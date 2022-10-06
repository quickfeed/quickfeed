package fileop

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// CopyDir copies the src directory to dst, ignoring .git directories.
func CopyDir(src, dst string) error {
	dst = filepath.Join(dst, filepath.Base(src))
	return filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.Contains(path, ".git") {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(dstPath), 0o700); err != nil {
			return err
		}
		return copyFile(path, dstPath)
	})
}

// copyFile copies srcFile to dstFile.
func copyFile(srcFile, dstFile string) (err error) {
	in, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := in.Close()
		if err == nil {
			err = closeErr
		}
	}()

	out, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := out.Close()
		if err == nil {
			err = closeErr
		}
	}()

	_, err = io.Copy(out, in)
	return
}
