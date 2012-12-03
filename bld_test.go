package main

import (
	"fmt"
	"regexp"
	"testing"
)

func TestError(test *testing.T) {
	defer catch()

	func() {
		defer rethrow("f1 %s", "arg")

		func() {
			throw("f2 %s", "arg")
		}()
	}()
}

func TestMacro(test *testing.T) {
	def := defines{
		"name1": []string{"../test/x.cpp", "a.cpp", "b.cpp"},
		"name2": []string{"../../../dir1", "dir2"},
		"name3": []string{"name1", "name2"},
	}

	testPrintStrings(def.substituteUserDefs([]string{"zzz ${/name2}/${name1} vvv"}))

	fmt.Println("----")
	testPrintStrings(def.substituteUserDefs([]string{"${name3}/${/${name3}}"}))

	setCurrentFiles([]string{"file", "other"})
	setTargeDir("../dir")
	//testPrintStrings(substituteEmbDefs([]string{"${.}/${..}/${*}"}))
}

func testPrintStrings(s []string) {
	for _, v := range s {
		fmt.Println(v)
	}
}

func TestSearchIncludes(test *testing.T) {
	re := regexp.MustCompile(`#include [<|"]([a-zA-Z0-9\/\\\._\-]+)[>|"]`)

	testString := `

 #include <util/delay.h>

#include <pdefs.hpp>
#include <usart.hpp>
#include "user.hpp"
#include "power.hpp"
#include <timer.hpp>
#include <watchdog.hpp>
#include <soft_uart.h>
#include <flow.hpp>
#include <io.hpp>
#include <gsm.hpp>
#include <mcu.hpp>
#include <sim900.hpp>

#include <log.hpp>
#include <freqcal.hpp>

void Every10Ms() {
	User::Tick();
	Flow::Tick();
}

`

	found := re.FindAllStringSubmatch(testString, -1)
	test.Error(found)
}

func TestGetSourceDeps(test *testing.T) {

	path := "../../test/main.cpp"
	searchDirs := []string{
		"../../test/debug",
		"../../test/fmt",
		"../../test/gsm",
		"../../test/hal",
		"../../test/io",
		"../../test/test",
		"../../test/types",
	}

	deps := getSourceDeps(path, searchDirs)

	for _, d := range deps {
		test.Error(d)
	}
}

func TestCache(test *testing.T) {
	const cacheFileName = "../../test/cache.json"
	cache := ReadCache(cacheFileName)
	path := "../../test/main.cpp"
	searchDirs := []string{
		"../../test/debug",
		"../../test/fmt",
		"../../test/gsm",
		"../../test/hal",
		"../../test/io",
		"../../test/test",
		"../../test/types",
	}

	cache.CheckSource(path, searchDirs)
	cache.CheckSource(path, searchDirs)

	cache.Write(cacheFileName)
}
