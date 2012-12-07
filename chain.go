package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	//"regexp"
)

type Operation struct {
	Name  string   `json:"name"`
	Descr string   `json:"descr"`
	Deps  []string `json:"deps"`

	Sources string   `json:"sources"`
	Dirs    []string `json:"dirs"`

	Group bool `json:"group"`

	Tool string   `json:"tool"`
	Args []string `json:"args"`

	// Хранит закешированные опции, с подстановленными переменными, кроме {}.
	cachedOpts []string
	// Список обрабатываемых файлов
	targetFiles []string
}

// выполняет операцию
func (op *Operation) Exec() {
	log.Printf("exec %s for %v with %v\n",
		op.Tool, op.targetFiles, op.cachedOpts)

	if op.Group {
		if len(op.targetFiles) == 0 {
			return
		}

		opts := substituteEmbDefs(op.cachedOpts, op.targetFiles)

		execCommand(op.Tool, opts)
		return
	}

	for _, file := range op.targetFiles {
		opts := substituteEmbDefs(op.cachedOpts, []string{file})
		execCommand(op.Tool, opts)
	}
}

// execCommand вызывает утилиту с указанными параметрами
func execCommand(cmd string, opts []string) {
	defer rethrow("command %s exec", cmd)

	run := exec.Command(cmd, opts...)
	out, err := run.CombinedOutput()
	os.Stdout.Write(out)

	if err != nil {
		panic(err)
	}
}

// Составляет список обрабатываемых файлов.
// dirs, root и targ должны содержать полные пути.
func (ci *Operation) SearchFiles(root, targ string, cache fileCache, defs defines) {

	if len(ci.Dirs) == 0 {
		ci.Dirs = []string{targ}
	} else {
		ci.Dirs = defs.substituteUserDefs(ci.Dirs)
	}

	// построение списка файлов
	changed := false
	ci.targetFiles = make([]string, 0, 32)
	for _, dir := range ci.Dirs {

		f, err := os.Open(dir)
		if err != nil {
			throw("Source dir not found: %s", err)
		}
		defer f.Close()

		finfs, err := f.Readdir(-1)
		if err != nil {
			panic(err)
		}

		for _, fi := range finfs {

			name := filepath.Join(dir, fi.Name())

			// проверка совпадения имени
			if isNameMatch(name, ci.Sources) {
				// проверка изменен ли файл
				if ci.Group {
					if cache.CheckSource(name, ci.Dirs) {
						changed = true
					}
					ci.targetFiles = append(ci.targetFiles, name)
				} else {
					if cache.CheckSource(name, ci.Dirs) {
						ci.targetFiles = append(ci.targetFiles, name)
					}
				}
			}
		}
	}
	if ci.Group && !changed {
		ci.targetFiles = []string{}
	}
}

// Возвращает true, если имя совпадает с паттерном
func isNameMatch(name string, exts string) bool {
	return filepath.Ext(name) == exts
}

// Кэширует опции в поле cachedOpt,
// подставляя указанные переменные.
func (ci *Operation) CacheOpts(defs defines) {
	ci.cachedOpts = defs.substituteUserDefs(ci.Args)
}

func (ci *Operation) Out() {
	printArr := func(descr string, a []string) {
		fmt.Println(descr)
		for _, s := range a {
			fmt.Print(s)
			fmt.Print("  ")
		}
		fmt.Println()
	}

	printArr("SOURCE_DIRS:", ci.Dirs)
	printArr("TARGET_FILES:", ci.targetFiles)

	printArr("OPTIONS:", ci.Args)
	printArr("CACHED_OPTS:", ci.cachedOpts)
}
