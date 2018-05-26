package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/kennygrant/sanitize"

	"github.com/dustin/go-wikiparse"
)

const indexFileSuffix = "multistream-index.txt.bz2"
const dataFileSuffix = "multistream.xml.bz2"

func readDump(indexFile string) error {

	if !strings.HasSuffix(indexFile, indexFileSuffix) {

		return fmt.Errorf("Wrong filename for indexfile. Expected '*%s', got %q", indexFileSuffix, indexFile)
	}
	dataFile := strings.TrimSuffix(indexFile, indexFileSuffix) + dataFileSuffix

	p, err := wikiparse.NewIndexedParser(indexFile, dataFile, runtime.NumCPU())
	if err != nil {
		return err
	}

	err = os.RemoveAll("out")
	if err != nil {
		return err
	}
	err = os.MkdirAll("out", 0777)
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go writeAllPagesToFiles(p, &wg)
	}
	wg.Wait()

	return nil
}

func writeAllPagesToFiles(p wikiparse.Parser, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		page, err := p.Next()
		if err == nil {
			err := writePageToFile("out", page)
			if err != nil {
				fmt.Println(err.Error())
			}
		} else {
			break
		}
	}

}

func writePageToFile(outDir string, page *wikiparse.Page) error {
	var prefixlen = 20
	if strlen := len(page.Revisions[0].Text); strlen < 20 {
		prefixlen = strlen
	}
	prefix := page.Revisions[0].Text[0:prefixlen]
	prefix = strings.ToLower(prefix)
	if strings.HasPrefix(prefix, "#redirect") {
		return nil
	}
	if strings.HasPrefix(prefix, "#weiterleitung") {
		return nil
	}

	baseName := sanitize.BaseName(page.Title)
	if len(baseName) == 0 {
		baseName = "_empty"
	}
	fileName := outDir + "/" + baseName + ".txt"
	i := 0
	for {
		_, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			break
		}
		fileName = outDir + "/" + fmt.Sprintf("%s-%d.txt", baseName, i)
		i++
	}

	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.WriteString(page.Title + "\n")
	if err != nil {
		return err
	}

	_, err = out.WriteString(page.Revisions[0].Text)
	if err != nil {
		return err
	}

	return nil
}
