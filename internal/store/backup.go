package store

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
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
	f.Close()
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
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return err
	}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		outPath := filepath.Join(dataDir, f.Name)
		rc, err := f.Open()
		if err != nil {
			return err
		}
		out, err := os.Create(outPath)
		if err != nil {
			rc.Close()
			return err
		}
		if _, err := io.Copy(out, rc); err != nil {
			rc.Close()
			out.Close()
			return err
		}
		rc.Close()
		out.Close()
	}
	return nil
}
