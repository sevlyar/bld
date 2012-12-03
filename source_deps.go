package main

import (
	"io/ioutil"
	//"log"
	"os"
	"path/filepath"
	"regexp"
)

// getSourceDepes возвращает список путей к файлам-зависимостям для
// указанного файла. Все пути указываются относительно рабочей директории.
func getSourceDeps(path string, searchDirs []string) []string {
	defer rethrow("unable get dependencies for source file %s", path)

	ext := filepath.Ext(path)
	if prov, ok := depsProviders[ext]; ok {
		return searchDeps(path, prov, searchDirs)
	}

	return make([]string, 0)
}

// depsSearch обходит по дереву зависимостей для указанного файла и
// строит список зависимостей.
func searchDeps(path string, prov depsProvider, searchDirs []string) []string {
	known := searchFiles(prov(path), searchDirs)

	for i := 0; i < len(known); i++ {

		// для каждого уже известного файла-зависимости
		// получение зависимостей
		deps := searchFiles(prov(known[i]), searchDirs)
		for _, d := range deps {

			// добавление в список известных, если отсутствует
			var exists bool
			for k, _ := range known {
				if exists = known[k] == d; exists {
					break
				}
			}
			if !exists {
				known = append(known, d)
			}
		}
	}

	return known
}

// searchFiles ищет файлы с указанными именами в указанных директориях, 
// заменяет имена путями к файлам относительно рабочей директории.
func searchFiles(names, dirs []string) []string {

	last := 0
	for j := 0; j < len(names); j++ {

		path := searchFile(names[j], dirs)

		if len(path) > 0 {
			names[last] = path
			last++
		}
	}

	return names[:last]
}

// searchFile ищет файл с указанным именем в указанных директориях, 
// возвращает путь относительно рабочей директории к первому наденному
// или пустую строку, если не найден.
func searchFile(name string, dirs []string) (path string) {

	for _, dir := range dirs {

		path = filepath.Join(dir, name)
		if fi, _ := os.Stat(path); fi != nil {
			break
		}
		path = ""
	}

	return
}

// depsProvider является сигнатурой для функций,
// реализующих получение из исходного файла имен
// файлов-зависимостей.
type depsProvider func(path string) []string

// карта, отображающая расширение файла в функцию,
// извлекающую зависимости.
var depsProviders = map[string]depsProvider{
	".c":   clangSourceDepsProvider,
	".cpp": clangSourceDepsProvider,
	".h":   clangSourceDepsProvider,
	".hpp": clangSourceDepsProvider,
}

func clangSourceDepsProvider(path string) []string {
	return getRegIncl(path, clangIncludeRegexp)
}

var (
	clangIncludeRegexp = regexp.MustCompile(
		`#include [<|"]([a-zA-Z0-9\/\\\._\-]+)[>|"]`)
)

// Получает список подключаемых файлов по регулярному выражению.
func getRegIncl(path string, exp *regexp.Regexp) []string {

	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	inclBytes := exp.FindAllSubmatch(b, -1)

	ret := make([]string, len(inclBytes))
	for i, _ := range inclBytes {
		ret[i] = string(inclBytes[i][1])
	}

	return ret
}
