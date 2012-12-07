package main

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var testEnv = map[string][]byte{

	"main.c": []byte(`
#include <c.h>
#include <a.h>
void main() {}
`),

	"a.h": []byte(`
#include <b.h>
#include <c.h>
`),

	"a.c": []byte(`
#include <a.h>
void f() {};
`),

	"d/b.h": []byte(`
#include <c.h>
#define TEST 1
`),

	"d/b.c": []byte(`
#include <b.h>
`),

	"d/c.h": []byte(`
#include <x.h>
`),

	"macro.json": []byte(`
{
  	"defs": {
  		"-ALPHA": ["a", "b"],
  		"-NUMB": ["1", "2"],
  		"-AB": ["{-ALPHA}", "{-NUMB}"],
        "-TEST-DEF": ["${.}/${..}/${-ALPHA}/${-NUMB}", "$${-AB}-$${-AB}"]
    }
}
`),

	"list.json": []byte(`
{
	"combine": [ "macro.json" ],

	"ops": [
		{
	        "name": "list-files",
	        "descr": "list modified files",

	        "sources": ".c",
	        "dirs": ["${..}/${DIRS}"],

	        "tool": "echo",
	        "args": ["tool \"${@}\""]
		}
	]
}
`),

	"build.json": []byte(`
{
	"combine": [ "list.json" ],
  	"defs": { 
        "DIRS": [".", "d"]
    }
}
`),
}

func TestCreateEnvironment(test *testing.T) {
	const FATAL_MSG = "unable create test environment:"

	const src = "../../test/"
	const bin = "../../test/bin"

	// create project files
	for name, content := range testEnv {
		fullPath := filepath.Join(src, name)

		// create directories
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			test.Fatal(FATAL_MSG, err)
		}

		// write file
		if err := ioutil.WriteFile(fullPath, content, 0644); err != nil {
			test.Fatal(FATAL_MSG, err)
		}
	}

	// create bin
	if err := os.MkdirAll(bin, 0755); err != nil {
		test.Fatal(FATAL_MSG, err)
	}

	TestChdir(test)
}

func TestChdir(test *testing.T) {
	const bin = "../../test/bin"

	if err := os.Chdir(bin); err != nil {
		test.Fatal(err)
	}
}
