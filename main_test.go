package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFileExtension(t *testing.T) {
	cases := []struct {
		input, expected string
	}{
		{"file.mp3", "mp3"},
		{"archive.tar.gz", "gz"},
		{"no_extension", ""},
	}

	for _, c := range cases {
		got := getFileExtension(c.input)
		if got != c.expected {
			t.Errorf("expected %s, got %s", c.expected, got)
		}
	}
}

func TestListFiles(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "test.txt"), []byte("dummy"), 0644)

	files, err := listFiles(dir)
	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 || files[0] != "test.txt" {
		t.Errorf("expected [test.txt], got %v", files)
	}
}

func TestMoveFileToTypeDir(t *testing.T) {
	tmp := t.TempDir()

	files := []string{"song.mp3", "photo.jpg", "note.txt", "strange.xyz"}
	for _, file := range files {
		err := os.WriteFile(filepath.Join(tmp, file), []byte("test"), 0644)
		if err != nil {
			t.Fatalf("cannot create test file %s: %v", file, err)
		}
	}

	analyzer := fileTypeAnalyzer{
		wd: tmp,
		mapper: map[string][]string{
			"music":  {"mp3"},
			"images": {"jpg"},
			"docs":   {"txt"},
		},
	}

	// Создаем директории заранее, как в реальном сценарии
	for dir := range analyzer.mapper {
		err := os.Mkdir(filepath.Join(tmp, dir), 0755)
		if err != nil {
			t.Fatalf("cannot create directory %s: %v", dir, err)
		}
	}

	err := analyzer.moveFileToTypeDir(files...)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Проверка
	cases := map[string]string{
		"song.mp3":    "music",
		"photo.jpg":   "images",
		"note.txt":    "docs",
		"strange.xyz": "misc",
	}

	for filename, expectedDir := range cases {
		fullPath := filepath.Join(tmp, expectedDir, filename)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("expected file %s to be in %s, but not found", filename, expectedDir)
		}
	}
}
