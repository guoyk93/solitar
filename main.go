package main

import (
	"encoding/json"
	"flag"
	"github.com/guoyk93/rg"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

type DatabaseBundle struct {
	Year string `json:"year"`
	Name string `json:"name"`
	Tape bool   `json:"tape"`
}

type Database struct {
	Bundles []*DatabaseBundle `json:"bundles"`
}

var (
	regexpYear = regexp.MustCompile(`^\d{4}$`)
)

var (
	ignoredDirs = map[string]struct{}{
		"@eaDir": {},
	}
)

var opts struct {
	cmdMigrate bool
}

func main() {
	var err error
	defer func() {
		if err == nil {
			return
		}
		log.Println("exited with error:", err.Error())
		os.Exit(1)
	}()
	defer rg.Guard(&err)

	flag.BoolVar(&opts.cmdMigrate, "migrate", false, "migrate archive json files")
	flag.Parse()

	if opts.cmdMigrate {
		doMigrate()
	}
}

func doMigrate() {
	var (
		db        Database
		fileJSONs []string
	)

	for _, entryYear := range rg.Must(os.ReadDir(".")) {
		if !entryYear.IsDir() {
			continue
		}
		if !regexpYear.MatchString(entryYear.Name()) {
			continue
		}

		for _, entryBundle := range rg.Must(os.ReadDir(entryYear.Name())) {
			if !entryBundle.IsDir() {
				continue
			}
			if _, ok := ignoredDirs[entryBundle.Name()]; ok {
				continue
			}

			fileBundleJSON := filepath.Join(entryYear.Name(), entryBundle.Name()+".json")

			buf, _ := os.ReadFile(fileBundleJSON)

			tape := false

			if len(buf) > 0 {
				var data struct {
					Tape bool `json:"tape"`
				}

				rg.Must0(json.Unmarshal(buf, &data))

				tape = data.Tape

				fileJSONs = append(fileJSONs, fileBundleJSON)
			}

			db.Bundles = append(db.Bundles, &DatabaseBundle{
				Year: entryYear.Name(),
				Name: entryBundle.Name(),
				Tape: tape,
			})
		}
	}

	rg.Must0(os.WriteFile("data.json", rg.Must(json.MarshalIndent(db, "", "  ")), 0644))

	for _, fileJSON := range fileJSONs {
		rg.Must0(os.Remove(fileJSON))
	}
}
