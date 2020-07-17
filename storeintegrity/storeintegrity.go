package storeintegrity

import (
	"fmt"
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/juju/errors"
	"github.com/mattn/go-zglob"
)

// StoreIntegrity checks the data store's pacakges integrity.
type StoreIntegrity struct {
	path string
}

// GetPath gets path.
func (i *StoreIntegrity) GetPath() string {
	return i.path
}

// SetPath sets path.
func (i *StoreIntegrity) SetPath(path string) {
	i.path = path
}

func (i *StoreIntegrity) Check() error {
	checkBluepint := `func \(\w+ \*\w+\) Table%s\(\) string \{\n\s+return "(\w+)"\n\}`
	tableNameCheck := regexp.MustCompile(fmt.Sprintf(checkBluepint, "Name"))
	tableAliasCheck := regexp.MustCompile(fmt.Sprintf(checkBluepint, "Alias"))
	entries, err := zglob.Glob(i.path + "/**/*.go")
	if err != nil {
		return errors.Trace(err)
	}

	tableNames := map[string]bool{}
	tableAliases := map[string]bool{}

	for _, entry := range entries {
		relativePath := strings.Replace(entry, i.path+"/", "", 1)
		chunks := strings.Split(relativePath, "/")
		fileName := chunks[len(chunks)-1]

		if fileName == "store.go" || strings.HasSuffix(fileName, "_test.go") {
			continue
		}

		b, err := ioutil.ReadFile(entry)
		if err != nil {
			return errors.Trace(err)
		}

		tableName := tableNameCheck.FindSubmatch(b)
		if len(tableName) > 0 {
			if _, ok := tableNames[string(tableName[1])]; ok {
				return errors.Errorf(fmt.Sprintf("Found duplicate table name `%s`", string(tableName[1])))
			}
			tableNames[string(tableName[1])] = true
		}

		tableAlias := tableAliasCheck.FindSubmatch(b)
		if len(tableAlias) > 0 {
			if _, ok := tableAliases[string(tableAlias[1])]; ok {
				return errors.Errorf(fmt.Sprintf("Found duplicate table alias `%s`", string(tableAlias[1])))
			}
			tableAliases[string(tableAlias[1])] = true
		}
	}
	return nil
}

// New returns a new instance of StoreIntegrity.
func New() (*StoreIntegrity, error) {
	i := &StoreIntegrity{}
	return i, nil
}
