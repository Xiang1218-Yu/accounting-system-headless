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
// 备份包来自外部（如客户回传），其条目名可能携带 "../" 或绝对路径，
// 直接拼接到 dataDir 会写出数据目录之外的文件（Zip Slip）。这里对每个条目
// 校验其相对 dataDir 解析后的路径必须落在目录内部，否则拒绝整包恢复。
func Restore(zipPath, dataDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return err
	}
	defer r.Close()
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(absDataDir, 0o755); err != nil {
		return err
	}
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		// 用方向斜杠统一后再 Join，避免条目名以 "\" 起头被 Windows 当作绝对路径。
		name := filepath.FromSlash(f.Name)
		if filepath.IsAbs(name) {
			return fmt.Errorf("restore: 拒绝越界条目 %q（绝对路径）", f.Name)
		}
		outPath := filepath.Join(absDataDir, name)
		rel, err := filepath.Rel(absDataDir, outPath)
		if err != nil || strings.HasPrefix(filepath.ToSlash(rel), "../") || rel == ".." {
			return fmt.Errorf("restore: 拒绝越界条目 %q", f.Name)
		}
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
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
