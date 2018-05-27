package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"sync"

	"github.com/dustin/go-wikiparse"
	"github.com/kennygrant/sanitize"
)

const indexFileSuffix = "multistream-index.txt.bz2"
const dataFileSuffix = "multistream.xml.bz2"

type DumpReader struct {
	Compress bool
	OutDir   string
	Parser   wikiparse.Parser
}

func (dr *DumpReader) readDump(indexFile string) error {

	if !strings.HasSuffix(indexFile, indexFileSuffix) {

		return fmt.Errorf("Wrong filename for indexfile. Expected '*%s', got %q", indexFileSuffix, indexFile)
	}
	dataFile := strings.TrimSuffix(indexFile, indexFileSuffix) + dataFileSuffix

	var err error = nil
	dr.Parser, err = wikiparse.NewIndexedParser(indexFile, dataFile, runtime.NumCPU())
	if err != nil {
		return err
	}

	err = os.RemoveAll(dr.OutDir)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	for i := 0; i < runtime.NumCPU(); i++ {
		wg.Add(1)
		go dr.writeAllPagesToFiles(&wg)
	}
	wg.Wait()

	return nil
}

func (dr *DumpReader) writeAllPagesToFiles(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		page, err := dr.Parser.Next()
		if err == nil {
			err := dr.writePageToFile(page)
			if err != nil {
				fmt.Println(err.Error())
			}
		} else {
			break
		}
	}

}

func (dr *DumpReader) writePageToFile(page *wikiparse.Page) error {
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

	out, err := dr.createOutputFile(page.Title)
	if err != nil {
		return err
	}
	defer out.Close()

	if dr.Compress {
		writer, err := gzip.NewWriterLevel(out, 3)
		if err != nil {
			return err
		}
		defer writer.Close()
		out = writer
	}

	_, err = io.WriteString(out, page.Title+"\n")
	if err != nil {
		return err
	}

	_, err = io.WriteString(out, page.Revisions[0].Text)
	if err != nil {
		return err
	}

	return nil
}

func (dr *DumpReader) createOutputFile(title string) (io.WriteCloser, error) {
	extension := ".txt"
	if dr.Compress {
		extension += ".gz"
	}
	baseName := sanitize.BaseName(title)
	baseName = strings.ToLower(baseName)
	if len(baseName) == 0 {
		baseName = "_empty"
	}

	dir := path.Join(dr.OutDir, baseName[0:1])
	err := os.MkdirAll(dir, 0777)
	if err != nil {
		return nil, err
	}

	fileName := path.Join(dir, baseName+extension)
	i := 0
	for {
		_, err := os.Stat(fileName)
		if os.IsNotExist(err) {
			break
		}
		fileName = path.Join(dir, fmt.Sprintf("%s-%d%s", baseName, i, extension))
		i++
	}

	out, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}
	return out, nil
}
