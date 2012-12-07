package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	path    string       `json:"-"`
	Combine []string     `json:"combine,omitempty"`
	Defs    defines      `json:"defs,omitempty"`
	Ops     []*Operation `json:"ops,omitempty"`
}

// loadConfigs загружает конфигурацию: читает указанный 
// конфигурационный файл и комбинирует его с необходимыми.
func loadConfigs(path string, dir string) *Config {
	defer rethrow("unable load configuration")

	root := readConfigFile(filepath.Join(dir, path))

	for i := 0; i < len(root.Combine); i++ {
		cnf := readConfigFile(filepath.Join(dir, root.Combine[i]))
		root.combine(cnf)
	}

	// check name uniq
	for i, op := range root.Ops {
		for j := i + 1; j < len(root.Ops); j++ {
			if op.Name == root.Ops[j].Name {
				throw("two or more operations has same name '%s'", op.Name)
			}
		}
	}

	log.Println("configuration loaded")
	return root
}

// readConfigFile читает и парсит указанный конфигурационный файл.
func readConfigFile(path string) *Config {
	log.Printf("loading config %s\n", path)

	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	conf := new(Config)
	err = json.NewDecoder(f).Decode(conf)
	if err != nil {
		panic(err)
	}

	conf.path = path
	return conf
}

// combine присоединяет к конфигурации root конфигурацию cnf 
// по следующим правилам:
//	* Элементы списка Combine из cnf, отсутствующие в root добавляются в конец
//	аналогичного списка;
//	* Списки Defs объединяются, значения определений с одинаковыми именами
//	объединяются;
//	* Списки Ops объединяются, не допускается совпадение имен (проверяется в
// 	процедуре загрузки).
func (root *Config) combine(cnf *Config) {
	log.Printf("combine config %s", cnf.path)

	for _, s := range cnf.Combine {
		if stringIndex(root.Combine, s) < 0 {
			root.Combine = append(root.Combine, s)
		}
	}

	for key, v := range cnf.Defs {
		if rootV, exists := root.Defs[key]; exists {
			v = append(rootV, v...)
		}
		root.Defs[key] = v
	}

	root.Ops = append(root.Ops, cnf.Ops...)
}

// store сохраняет конфигурацию в json-файл (предназначена для диагностики).
func (conf *Config) store(path string) {
	defer rethrow("unable store configuration")

	body, err := json.MarshalIndent(conf, "", "\t")
	if err != nil {
		panic(err)
	}

	if err = ioutil.WriteFile(path, body, 0644); err != nil {
		panic(err)
	}

	log.Printf("config %s stored\n", path)
}
