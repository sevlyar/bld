package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	//"path/filepath"
)

// Конфигурация - глобальные настроки.
type Config struct {
	// Максимальное количество одновременно выполняемых процессов.
	Parallel int `json:"parallel"`
	// Список переменных конфигурации, применяются в опциях процесса.
	Defs defines `json:"defs"`
	// Цепочка процессов (Этапов) обработки.
	Chain []ChainItem `json:"chain"`
}

func fatal(a ...interface{}) {
	fmt.Println(a...)
	os.Exit(1)
}

func check(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
		os.Exit(1)
	}
}

func main() {
	// чтение опций
	// обработка каждого элемента цепочки:
	// подстановка глобальных переменных
	// поиск подходящих файлов
	// исключение неизменившихся + обновление кэша
	// для каждого найденого
	// инстанцирование имени файла
	// вызов инструмента

	// в текущей директории находится конфиг и кеш, она является целевой
	// root проекта указывается в параметре

	// чтение конфига
	conf := readConf()

	// получение целевой директории и директории проекта
	targetDir := "./"

	flag.Parse()
	rootDir := flag.Arg(0)

	setTargeDir(targetDir)
	setRootDir(rootDir)

	// разворачивание макроопределений
	conf.Defs.bootstrap()

	//fmt.Println(conf.Defs)

	// чтение кеша
	cache := ReadCache("gomakecache.json")

	// обработка цепочки
	for _, item := range conf.Chain {
		fmt.Println(item.Descr)
		item.SearchFiles(rootDir, targetDir, cache, conf.Defs)
		item.CacheOpts(conf.Defs)
		item.Exec()
	}

	// запись кеша
	cache.Write("gomakecache.json")
}

func readConf() *Config {
	f, err := os.Open("gomake.json")
	check("Unable open config file:", err)
	defer f.Close()

	conf := new(Config)
	err = json.NewDecoder(f).Decode(conf)
	check("Wrong config:", err)

	return conf
}
