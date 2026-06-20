package ferry

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func copyFile(src, dest string, fileMode, dirMode fs.FileMode, uid, gid int) error {
	if err := mkdirAllOwned(filepath.Dir(dest), dirMode, uid, gid); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()

	tmp, err := os.CreateTemp(filepath.Dir(dest), ".fotoferry-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(tmp.Name()) }()

	if _, err := io.Copy(tmp, in); err != nil {
		_ = tmp.Close()
		return err
	}
	// CreateTemp makes the file 0600; set the configured mode so the
	// destination is readable (e.g. over SMB by a non-root user).
	if err := tmp.Chmod(fileMode); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if uid >= 0 && gid >= 0 {
		if err := os.Chown(tmp.Name(), uid, gid); err != nil {
			return err
		}
	}
	return os.Rename(tmp.Name(), dest)
}

// mkdirAllOwned is os.MkdirAll but it forces the exact mode (ignoring umask) and
// chowns each directory it creates, so created folders match the destination's
// ownership convention.
func mkdirAllOwned(path string, mode fs.FileMode, uid, gid int) error {
	if fi, err := os.Stat(path); err == nil {
		if !fi.IsDir() {
			return fmt.Errorf("%s exists and is not a directory", path)
		}
		return nil
	}

	parent := filepath.Dir(path)
	if parent != path {
		if err := mkdirAllOwned(parent, mode, uid, gid); err != nil {
			return err
		}
	}

	if err := os.Mkdir(path, mode); err != nil && !os.IsExist(err) {
		return err
	}
	if err := os.Chmod(path, mode); err != nil {
		return err
	}
	if uid >= 0 && gid >= 0 {
		if err := os.Chown(path, uid, gid); err != nil {
			return err
		}
	}
	return nil
}
