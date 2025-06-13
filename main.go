package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type fileAnalyzer interface {
	analyzeAndSort() error
}

func analyze(fa fileAnalyzer) error {
	return fa.analyzeAndSort()
}

var blacklist = []string{
	"go",
	"mod",
	"exe",
}

func getFileExtension(name string) string {
	fmt.Println("getfile")
	return strings.TrimPrefix(filepath.Ext(name), ".")

}

func listFiles(dirname string) ([]string, error) {
	//creating a file for locate in the dirnames
	var files []string
	fmt.Println("listfiles")

	//assignment to a variable list of files in dirname
	list, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	//appending files array list of files in dirname
	for _, file := range list {
		if !file.IsDir() {
			files = append(files, file.Name())
		}
	}

	return files, nil
}

func listDirs(dirname string) ([]string, error) {
	var dirs []string
	fmt.Println("listdirs")

	list, err := os.ReadDir(dirname)
	if err != nil {
		return nil, err
	}

	for _, file := range list {
		if file.IsDir() {
			dirs = append(dirs, file.Name())
		}
	}

	return dirs, nil
}

func mkdir(dirname string) error {
	//creating directory
	err := os.Mkdir(dirname, 0644)
	fmt.Println("mkdirs")

	if err != nil && os.IsExist(err) {
		return nil
	}

	var e *os.PathError

	if err != nil && errors.As(err, &e) {
		return nil
	}

	return err
}

func moveFile(name string, dst string) error {
	fmt.Println("movefile")
	return os.Rename(name, filepath.Join(dst, name))
}

// func getCurrentDate(t time.Time) string {
// 	return t.Format("2005-01-02")
// }

func filter[T any](ts []T, fn func(T) bool) []T {
	filtered := make([]T, 0)
	fmt.Println("filter")

	for i := range ts {
		if fn(ts[i]) {
			filtered = append(filtered, ts[i])
		}
	}

	return filtered
}

type fileTypeAnalyzer struct {
	wd     string
	mapper map[string][]string
}

func newFileTypeAnalyzer(wd string) *fileTypeAnalyzer {
	fmt.Println("newfiletype")
	return &fileTypeAnalyzer{
		wd: wd,
		mapper: map[string][]string{
			"video":  {"mp4", "mkv", "3gp", "wmv", "flv", "avi", "mpeg", "webm"},
			"music":  {"mp3", "aac", "wav", "flac"},
			"images": {"jpg", "jpeg", "png", "gif", "svg", "tiff"},
			"docs":   {"docx", "csv", "txt", "xlsx"},
			"books":  {"pdf", "epub"},
		},
	}
}

func (f fileTypeAnalyzer) analyzeAndSort() error {
	files, err := listFiles(f.wd)
	fmt.Println("analyzeandsort")
	if err != nil {
		return fmt.Errorf("could not list files: %w", err)
	}

	if err := f.createFileTypeDirs(files...); err != nil {
		return err
	}

	return f.moveFileToTypeDir(files...)
}

func (f fileTypeAnalyzer) moveFileToTypeDir(files ...string) error {
	dirs, err := listDirs(f.wd)
	if err != nil {
		return fmt.Errorf("could not list directories: %w", err)
	}

	for _, dir := range dirs {
		for _, file := range files {
			if slices.Contains(f.mapper[dir], strings.ToLower(getFileExtension(file))) {
				if err := moveFile(file, dir); err != nil {
					return err
				}
			}
		}
	}

	files, err = listFiles(f.wd)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return nil
	}

	files = filter(files, func(f string) bool {
		return !slices.Contains(blacklist, getFileExtension(f))
	})

	for i := range files {
		if err := f.moveToMisc(files[i]); err != nil {
			return err
		}
	}

	return nil
}

func (f fileTypeAnalyzer) moveToMisc(file string) error {
	fmt.Println("movetomsic")
	if err := mkdir("misc"); err != nil {
		return err
	}

	return moveFile(file, "misc")
}

func (f fileTypeAnalyzer) createFileTypeDirs(files ...string) error {
	fmt.Println("createfiletypedirs")
	for ftype, fvalues := range f.mapper {
		for _, file := range files {
			if slices.Contains(fvalues, getFileExtension(file)) {
				if err := mkdir(ftype); err != nil {
					return fmt.Errorf("could not create folder: %w", err)
				}
			}
		}
	}

	return nil
}

func main() {
	logger := slog.Default()
	slog.SetDefault(logger)

	var analyzer fileAnalyzer

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var mode string

	flag.StringVar(&mode, "mode", "", "Provide sort mode (type|data)")
	flag.Parse()

	switch mode {
	case "type":
		analyzer = newFileTypeAnalyzer(wd)
	default:
		fmt.Println("Provide sort mode flag: --mode=(type|date)")
		return
	}

	if err := analyze(analyzer); err != nil {
		log.Fatal(err)
	}
}
