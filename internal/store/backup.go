package store

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Backup 将数据目录所有文件打包到 zip
func (s *Store) Backup(dest string) error {
	tmp := dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	defer f.Close()

	w := zip.NewWriter(f)
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if err := addFileToZip(w, filepath.Join(s.dir, e.Name()), e.Name()); err != nil {
			w.Close()
			return err
		}
	}
	if err := w.Close(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, dest)
}

func addFileToZip(w *zip.Writer, src, name string) error {
	srcF, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcF.Close()
	info, err := srcF.Stat()
	if err != nil {
		return err
	}
	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	hdr.Name = name
	hdr.Method = zip.Deflate
	dst, err := w.CreateHeader(hdr)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, srcF)
	return err
}

// Restore 从 zip 恢复数据目录（覆盖现有文件）。调用方需重新 New 重载。
func Restore(zipPath, dataDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()

	root, err := filepath.Abs(dataDir)
	if err != nil {
		return fmt.Errorf("解析数据目录: %w", err)
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return err
	}

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		outPath, err := restoreTarget(root, f)
		if err != nil {
			return err
		}
		if err := extractZipFile(f, outPath); err != nil {
			return err
		}
	}
	return nil
}

func restoreTarget(root string, f *zip.File) (string, error) {
	name := filepath.Clean(f.Name)
	if name == "." || name == "" || filepath.IsAbs(name) || filepath.VolumeName(name) != "" {
		return "", fmt.Errorf("备份条目路径无效: %q", f.Name)
	}
	if name == ".." || strings.HasPrefix(name, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("备份条目越出数据目录: %q", f.Name)
	}
	target, err := filepath.Abs(filepath.Join(root, name))
	if err != nil {
		return "", fmt.Errorf("解析备份条目: %w", err)
	}
	relative, err := filepath.Rel(root, target)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("备份条目越出数据目录: %q", f.Name)
	}
	if !f.Mode().IsRegular() {
		return "", fmt.Errorf("备份条目不是普通文件: %q", f.Name)
	}
	return target, nil
}

func extractZipFile(f *zip.File, outPath string) error {
	source, err := f.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	mode := f.Mode().Perm()
	if mode == 0 {
		mode = 0o644
	}
	destination, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(destination, source)
	closeErr := destination.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return nil
}
