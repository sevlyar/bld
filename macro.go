package main

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Макрос ${name}
// Предопределенные макросы
// ${.} - целевой каталог
// ${..} - каталог проекта
// ${@} - подставляется полное имя обрабатываемого файла
// ${*} - подставляются имена всех обрабатываемых файлов в текущем звене
// ${/name} - модификатор, указывающий, что надо взять базовое имя и трактовать макрос как путь

type defines map[string][]string

var (
	macroRegexp    = regexp.MustCompile(`\$\{\/?[^\.\{\}@]*\}`)
	preMacroRegexp = regexp.MustCompile(`\$\{\/?(\.\.|\.)\}`)
	embMacroRegexp = regexp.MustCompile(`\$\{\/?(\.\.|\.|@)\}`)

	embeddedDefs = defines{
		"..": []string{".."},
		".":  []string{"."},
		"@":  []string{"?"},
		//"*":  []string{"?"},
	}
)

func setRootDir(path string) {
	embeddedDefs[".."][0] = path
}

func setTargeDir(path string) {
	embeddedDefs["."][0] = path
}

func setCurrentFiles(files []string) {
	embeddedDefs["@"] = files
}

// func setCurrentFiles(files []string) {
// 	embeddedDefs["*"] = files
// }

// Раскручивает определения, делая всевозможные подстановки.
func (def defines) bootstrap() {
	for name, values := range def {
		res := make([]string, 0, 64)
		for _, val := range values {
			res = append(res, def.substitute([]string{val}, macroRegexp)...)
		}
		def[name] = substituteEmbDefs(res, true)
	}
}

func substituteEmbDefs(input []string, pre bool) []string {
	re := embMacroRegexp
	if pre {
		re = preMacroRegexp
	}

	res := make([]string, 0, 64)
	for _, val := range input {
		res = append(res, embeddedDefs.substitute([]string{val}, re)...)
	}
	return res
}

func (def defines) substituteUserDefs(input []string) []string {
	res := make([]string, 0, 64)
	for _, val := range input {
		res = append(res, def.substitute([]string{val}, macroRegexp)...)
	}
	return res
}

// Делает подстановки определений на места макровызовов.
// input должен содержать одинаковые макровызовы!
// FIXME: сделать проверку зацикливания.
func (def defines) substitute(input []string, re *regexp.Regexp) []string {
	for {
		// выходные значения на каждой итерации
		result := make([]string, 0, 16)

		for _, str := range input {
			// поиск макровызова
			macroCall := re.FindString(str)

			if len(macroCall) == 0 {
				result = append(result, str)
				return input
			}

			// извлечение имени
			name := macroCall[2 : len(macroCall)-1]

			// проверка префикса
			basePath := false
			if name[0] == '/' {
				basePath = true
				name = name[1:]
			}

			// поиск значения
			values, exists := def[name]
			if !exists {
				throw("Macro definition with name %s not found", name)
			}

			// применение модификатора
			if basePath {
				values = basePathModif(values)
			}

			// подстановка каждым значением макроопределения
			for _, val := range values {
				subs := strings.Replace(str, macroCall, val, 1)
				result = append(result, subs)
			}
		}

		input = result
	}

	return input
}

// Модификатор значений переменной - базовый путь.
func basePathModif(v []string) []string {
	short := make([]string, len(v))
	for i, _ := range v {
		short[i] = filepath.Base(v[i])
	}
	return short
}
