package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"github.com/mattn/go-zglob"
)

func checkStoresIntegrity() {
	checkBluepint := `func \(\w+ \*\w+\) Table%s\(\) string \{\n\s+return "(\w+)"\n\}`
	tableNameCheck := regexp.MustCompile(fmt.Sprintf(checkBluepint, "Name"))
	tableAliasCheck := regexp.MustCompile(fmt.Sprintf(checkBluepint, "Alias"))
	storesPath := "./stores"
	entries, err := zglob.Glob(storesPath + "/**/*.go")
	if err != nil {
		log.Fatal(err)
	}

	tableNames := map[string]bool{}
	tableAliases := map[string]bool{}

	for _, entry := range entries {
		relativePath := strings.Replace(entry, storesPath+"/", "", 1)
		chunks := strings.Split(relativePath, "/")
		fileName := chunks[len(chunks)-1]

		if fileName == "store.go" || strings.HasSuffix(fileName, "_test.go") {
			continue
		}

		b, err := ioutil.ReadFile(entry)
		if err != nil {
			log.Fatal(err)
		}

		tableName := tableNameCheck.FindSubmatch(b)
		if len(tableName) > 0 {
			if _, ok := tableNames[string(tableName[1])]; ok {
				log.Fatal(fmt.Sprintf("Found duplicate table name `%s`", string(tableName[1])))
			}
			tableNames[string(tableName[1])] = true
		}

		tableAlias := tableAliasCheck.FindSubmatch(b)
		if len(tableAlias) > 0 {
			if _, ok := tableAliases[string(tableAlias[1])]; ok {
				log.Fatal(fmt.Sprintf("Found duplicate table alias `%s`", string(tableAlias[1])))
			}
			tableAliases[string(tableAlias[1])] = true
		}
	}
}
