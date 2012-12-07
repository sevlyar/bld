package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

var (
	verbose    bool
	macroLevel int
)

func init() {
	const (
		usage_verbose    = "enable verbose output"
		usage_macroLevel = "max level of macro"
	)

	flag.BoolVar(&verbose, "-verbose", false, usage_verbose)
	flag.BoolVar(&verbose, "v", false, usage_verbose)

	flag.IntVar(&macroLevel, "-level", 9, usage_macroLevel)
	flag.IntVar(&macroLevel, "l", 9, usage_macroLevel)
}

// TODO: организовать хранение в репозитарии и обозначить направление развития
// TODO: сделать проверку зависимостей и построение с учетом зависимостей
// TODO: реализовать макрос ${#}
func main() {
	defer exit()

	flag.Parse()

	if verbose {
		log.SetFlags(log.Lmicroseconds)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	scenario := "build.json"
	root := ".."
	if flag.NArg() > 0 {
		scenario = flag.Arg(0)
		if len(filepath.Ext(scenario)) == 0 {
			scenario += ".json"
		}
		if flag.NArg() > 1 {
			root = flag.Arg(1)
		}
	}

	if wd, err := os.Getwd(); err != nil {
		panic(err)
	} else {
		log.Println("work dir: ", wd)
	}
	log.Println("root dir: ", root)
	log.Println("scenario: ", scenario)

	conf := loadConfigs(scenario, root)

	conf.Defs.set(".", ".")
	conf.Defs.set("..", root)
	conf.Defs.bootstrap()

	cache := ReadCache("gomakecache.json")

	for _, item := range conf.Ops {
		fmt.Println(item.Descr)
		item.SearchFiles(root, ".", cache, conf.Defs)
		item.CacheOpts(conf.Defs)
		item.Exec()
	}

	cache.Write("gomakecache.json")
}
