package main

import (
	"testing"
)

func TestConfig(test *testing.T) {
	defer exit()

	verbose = true
	TestCreateEnvironment(test)

	conf := loadConfigs("build.json", "..")
	conf.store("combined.json")
}

func TestMain(test *testing.T) {
	verbose = true

	TestChdir(test)
	main()
}
