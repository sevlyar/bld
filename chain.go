package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

type Operation struct {
	Name  string   `json:"name"`
	Descr string   `json:"descr"`
	Deps  []string `json:"deps"`

	Sources []string `json:"sources"`
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
func (op *Operation) SearchFiles(root, targ string, cache fileCache, defs defines) {

	if len(op.Dirs) == 0 {
		op.Dirs = []string{targ}
	} else {
		op.Dirs = defs.substituteUserDefs(op.Dirs)
	}
	
	op.Sources = defs.substituteUserDefs(op.Sources)

	// построение списка файлов
	changed := false
	op.targetFiles = make([]string, 0, 32)
	for _, dir := range op.Dirs {

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
			if isNameMatch(name, op.Sources) {
				// проверка изменен ли файл
				if op.Group {
					if cache.CheckSource(name, op.Dirs) {
						changed = true
					}
					op.targetFiles = append(op.targetFiles, name)
				} else {
					if cache.CheckSource(name, op.Dirs) {
						op.targetFiles = append(op.targetFiles, name)
					}
				}
			}
		}
	}
	if op.Group && !changed {
		op.targetFiles = []string{}
	}
}

// Возвращает true, если имя совпадает с одним из паттернов
func isNameMatch(name string, pats []string) bool {
	for _, pat := range pats {
		match, err := regexp.MatchString(pat, name)
		if err != nil {
			panic(err)
		}
		if match {
			return true
		}
	}
	return false
}

// Кэширует опции в поле cachedOpt,
// подставляя указанные переменные.
func (op *Operation) CacheOpts(defs defines) {
	op.cachedOpts = defs.substituteUserDefs(op.Args)
}

func (op *Operation) Out() {
	printArr := func(descr string, a []string) {
		fmt.Println(descr)
		for _, s := range a {
			fmt.Print(s)
			fmt.Print("  ")
		}
		fmt.Println()
	}

	printArr("SOURCE_DIRS:", op.Dirs)
	printArr("TARGET_FILES:", op.targetFiles)

	printArr("OPTIONS:", op.Args)
	printArr("CACHED_OPTS:", op.cachedOpts)
}
