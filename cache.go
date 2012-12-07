package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"io"
	//"io/ioutil"
	"log"
	"os"
	"time"
)

// Хранит информацию об атрибутах файлов, чтобы определить
// какие изменились с прошлого запуска.
type fileCache map[string]*FileStateSnap

// Кэш находится в рабочей директории, он хранит информацию о прошлом сотоянии
// обрабатываемых файлов и позволяет узнать, какие файлы или их зависимости
// изменились и требуют обработки.
// Кэш представлен ассоциативным контейнером, где ключем является путь к файлу
// относительно рабочей директории, а значением - атрибуты файла.
type FileStateSnap struct {
	Path string `json:"path"`
	// Указывает что файл был модифицирован со времени прошлой обработки
	Modified bool `json:"-"`
	// Время последней модификации во время прошлой обработки
	Time time.Time `json:"time"`
	// Хэш файла во время прошлой обработки
	Hash []byte `json:"hash"`
	// Зависимости файла, пути относительно рабочей директории
	Depends []string `json:"dependencies"`
}

// Создает снимок состояния файла.
func ShotFileState(path string, dirs []string) *FileStateSnap {
	snap := new(FileStateSnap)
	snap.Path = path
	snap.Sync(dirs)
	return snap
}

// Синхронизирует снимок с состоянием файла.
func (item *FileStateSnap) Sync(dirs []string) {
	item.Modified = true
	item.Time = getFileInfo(item.Path).ModTime()
	item.Hash = getFileHash(item.Path)
	item.Depends = getSourceDeps(item.Path, dirs)
}

// Чтение кэша из файла.
func ReadCache(path string) fileCache {
	defer rethrow("unable load cache %s", path)

	log.Printf("read cache %s\n", path)

	cache := make(map[string]*FileStateSnap)

	f, err := os.Open(path)
	if err == nil {
		defer f.Close()

		err = json.NewDecoder(f).Decode(&cache)
		if err != nil {
			panic(err)
		}
	} else {
		if !os.IsNotExist(err) {
			if err != nil {
				panic(err)
			}
		}
	}

	return cache
}

// Запись кэша в файл.
func (cache *fileCache) Write(path string) {
	defer rethrow("unable store cache %s", path)
	log.Printf("write cache %s\n", path)

	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}

	defer f.Close()

	b, err := json.MarshalIndent(&cache, "", "\t")
	if err != nil {
		panic(err)
	}

	_, err = f.Write(b)
	if err != nil {
		panic(err)
	}
}

// Проверяет менялся ли файл или его зависимости, попутно добавляет или
// обновляет снимки файлов.
func (cache *fileCache) CheckSource(path string, dirs []string) bool {
	f := cache.Check(path, dirs)
	for _, dep := range (*cache)[path].Depends {
		f = cache.Check(dep, dirs) || f
	}
	return f
}

// Ищет файл в кеше. Если находит проверяет изменился ли он, и возвращает 
// результат. Если не находит, добавляет его и зависимости в кэш, возвращает
// true.
func (cache *fileCache) Check(path string, dirs []string) bool {

	item, exists := (*cache)[path]

	switch {
	case !exists:
		log.Printf("placing in cache %s\n", path)

		snap := ShotFileState(path, dirs)
		(*cache)[path] = snap

		for _, d := range snap.Depends {
			cache.Check(d, dirs)
		}
		return true

	case item.Modified:
		log.Printf("file %s modified, allready in cache\n", path)
		return true

	case item.Time != getFileInfo(path).ModTime():
		log.Printf("file %s modified (time)\n", path)
		item.Sync(dirs)
		return true

	case !bytes.Equal(getFileHash(path), item.Hash):
		log.Printf("file %s modified (hash) %s\n", path)
		item.Sync(dirs)
		return true
	}

	log.Printf("file %s not modified\n", path)
	return false
}

func getFileInfo(path string) os.FileInfo {
	fi, err := os.Lstat(path)
	if err != nil {
		if !os.IsNotExist(err) {
			throw("Unable open file: %s", err)
		}
		return nil
	}
	return fi
}

func getFileHash(path string) []byte {
	defer rethrow("unable compute hash of file", path)

	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	Hash := md5.New()
	_, err = io.Copy(Hash, f)
	if err != nil {
		panic(err)
	}

	return Hash.Sum(nil)
}
