package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

// Описатель подключений внешних файлов для построения зависимостей
type IncludesDescr struct {
	// Директории, в которых ищутся подключаемые файлы
	FileDirs []string `json:"file-dirs"`
	// Регулярное выражение, описывающее подключение, должно иметь один
	// параметр, совпадающий с именем подключаемого файла:
	// `#include [<|"]([a-zA-Z0-9\/\\\._\-]+)[>|"]`
	RegExpStr string `json:"reg-exp"`

	regExp *regexp.Regexp
}

func (inc *IncludesDescr) getRegExp() *regexp.Regexp {
	defer rethrow("Wrong format includes description")

	if inc.regExp == nil {
		inc.regExp = regexp.MustCompile(inc.RegExpStr)
	}

	return inc.regExp
}

// Этап обработки - указывает действие, применяемое к группе файлов одного типа.
type ChainItem struct {
	Descr string `json:"descr"`
	// Перекрывает глобальные настроки директорий-источников.
	// Если список пустой, то файлы ищутся в целевой директории.
	SourceDirs []string `json:"source-dirs"`
	// Описание подключений для установления зависимостей между файлами
	Includes *IncludesDescr `json:"includes"`
	// Подставлять в имя сразу все файлы (tool вызывается один раз)
	Group bool `json:"group"`
	// Список расширений обрабатываемых файлов.
	FileExts []string `json:"file-exts"`
	// Инструмент, вызывается для каждого подходящего файла.
	Tool string `json:"tool"`
	// Список опций, в каждом элементе может быть указана одна переменная,
	// тогда при добавлении аргументов к вызову инструмента произойдет
	// инстанцировние шаблона аргумента столько раз, сколько элементов указано 
	// в значении переменной. {} указывает, что нужно подставить имя 
	// обрабатываемого файла.
	Opts []string `json:"options"`

	// Хранит закешированные опции, с подстановленными переменными, кроме {}.
	cachedOpts []string
	// Список обрабатываемых файлов
	targetFiles []string
}

func (ci *ChainItem) Exec() {

	if ci.Group {
		if len(ci.targetFiles) == 0 {
			return
		}

		setCurrentFiles(ci.targetFiles)
		opts := substituteEmbDefs(ci.cachedOpts, false)
		ci.execTool(opts)
		return
	}

	for _, file := range ci.targetFiles {

		setCurrentFiles([]string{file})
		opts := substituteEmbDefs(ci.cachedOpts, false)
		ci.execTool(opts)
	}
}

func (ci *ChainItem) execTool(opts []string) {
	cmd := exec.Command(ci.Tool, opts...)

	out, err := cmd.CombinedOutput()

	os.Stdout.Write(out)
	check("Tool return error:", err)
}

// Составляет список обрабатываемых файлов.
// dirs, root и targ должны содержать полные пути.
func (ci *ChainItem) SearchFiles(root, targ string, cache fileCache, defs defines) {

	if len(ci.SourceDirs) == 0 {
		ci.SourceDirs = []string{targ}
	} else {
		// подстановка макроопределений в пути
		ci.SourceDirs = defs.substituteUserDefs(ci.SourceDirs)
		ci.SourceDirs = substituteEmbDefs(ci.SourceDirs, true)
	}

	if ci.Includes != nil {
		ci.Includes.FileDirs = defs.substituteUserDefs(ci.Includes.FileDirs)
		ci.Includes.FileDirs = substituteEmbDefs(ci.Includes.FileDirs, true)
	}

	// построение списка файлов
	changed := false
	ci.targetFiles = make([]string, 0, 32)
	for _, dir := range ci.SourceDirs {

		f, err := os.Open(dir)
		check("Source dir not found:", err)
		defer f.Close()

		finfs, err := f.Readdir(-1)
		check("Internal error:", err)

		for _, fi := range finfs {

			name := filepath.Join(dir, fi.Name())

			// проверка совпадения имени
			if isNameMatch(name, ci.FileExts) {
				// проверка изменен ли файл
				if ci.Group {
					if cache.CheckSource(name, ci.Includes.FileDirs) {
						changed = true
					}
					ci.targetFiles = append(ci.targetFiles, name)
				} else {
					if cache.CheckSource(name, ci.Includes.FileDirs) {
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
func isNameMatch(name string, exts []string) bool {
	for _, ext := range exts {
		if filepath.Ext(name) == ext {
			return true
		}
	}
	return false
}

// Кэширует опции в поле cachedOpt,
// подставляя указанные переменные.
func (ci *ChainItem) CacheOpts(defs defines) {
	ci.cachedOpts = defs.substituteUserDefs(ci.Opts)
	ci.cachedOpts = substituteEmbDefs(ci.cachedOpts, true)
}

func (ci *ChainItem) Out() {
	printArr := func(descr string, a []string) {
		fmt.Println(descr)
		for _, s := range a {
			fmt.Print(s)
			fmt.Print("  ")
		}
		fmt.Println()
	}

	printArr("SOURCE_DIRS:", ci.SourceDirs)
	printArr("TARGET_FILES:", ci.targetFiles)

	printArr("OPTIONS:", ci.Opts)
	printArr("CACHED_OPTS:", ci.cachedOpts)
}
