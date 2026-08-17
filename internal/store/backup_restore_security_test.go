package store

import (
	"archive/zip"
	"os"
	"path/filepath"
	"testing"
)

func TestRestoreRejectsArchiveEntriesOutsideDataDirectory(t *testing.T) {
	root := t.TempDir()
	archive := filepath.Join(root, "restore.zip")
	escaped := filepath.Join(root, "escaped.json")

	out, err := os.Create(archive)
	if err != nil {
		t.Fatalf("create archive: %v", err)
	}
	writer := zip.NewWriter(out)
	entry, err := writer.Create("../escaped.json")
	if err != nil {
		t.Fatalf("create entry: %v", err)
	}
	if _, err := entry.Write([]byte(`{"tampered":true}`)); err != nil {
		t.Fatalf("write entry: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close archive: %v", err)
	}
	if err := out.Close(); err != nil {
		t.Fatalf("close file: %v", err)
	}

	if err := Restore(archive, filepath.Join(root, "data")); err == nil {
		t.Fatal("恢复包内含越界路径时应被拒绝")
	}
	if _, err := os.Stat(escaped); !os.IsNotExist(err) {
		t.Fatalf("恢复操作写出了数据目录: stat error=%v", err)
	}
}
