package main

import (
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type defines map[string][]string

var (
	envMacroRegexp = regexp.MustCompile(`\$\{.*?\}`)
	macroRegexp    = regexp.MustCompile(`\$\(\/?[^@]*?\)`)
	embMacroRegexp = regexp.MustCompile(`\$\(\/?@\)`)
)

// set устанавливает значение макроопределения или добавляет новое.
func (def defines) set(name, val string) {
	def[name] = []string{val}
}

// Раскручивает определения, делая всевозможные подстановки.
func (def defines) bootstrap() {
	// для каждого макроопределения
	for name, values := range def {
		res := make([]string, 0, 64)
		// для каждого значения в макроопределении
		for _, val := range values {
			// выполнить подстановку макровызовов
			newVal := def.substitute([]string{val}, macroRegexp)
			res = append(res, newVal...)
		}
		def[name] = res
	}
}

func substituteEmbDefs(input, sources []string) []string {
	emb := defines{"@": sources}
	return emb.substituteDefs(input, embMacroRegexp)
}

func (def defines) substituteUserDefs(input []string) []string {
	return def.substituteDefs(input, macroRegexp)
}

func (def defines) substituteDefs(input []string, r *regexp.Regexp) []string {
	res := make([]string, 0, 64)
	for _, val := range input {
		res = append(res, def.substitute([]string{val}, r)...)
	}
	return res
}

// Делает подстановки определений на места макровызовов.
// input должен содержать одинаковые макровызовы!
// FIXME: сделать input string
func (def defines) substitute(input []string, re *regexp.Regexp) []string {
	cached := input[0]
	level := 0

	for {
		// выходные значения на каждой итерации
		result := make([]string, 0, 16)

		f := false
		for _, str := range input {
			// поиск макровызова
			macroCall := re.FindString(str)

			if len(macroCall) != 0 {
				f = true
			} else {
				continue
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

		if !f {
			// expand enveronment
			for i, _ := range input {
				input[i] = expandEnvVars(input[i])
			}

			if len(input) == 0 || cached != input[0] {
				log.Printf("macro L%d: [%s] -> %v", level, cached, input)
			}

			return input
		}

		input = result
		level++

		if level > macroLevel {
			throw("cyclic reference in macro-call %s or depends", cached)
		}
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

func getEnvVar(macro string) string {
	v := macro[2 : len(macro)-1]
	value := os.Getenv(v)
	log.Printf("os env: %s => %s", v, value)
	return value
}

func expandEnvVars(s string) string {
	ex := envMacroRegexp.ReplaceAllStringFunc(s, getEnvVar)
	return ex
}
