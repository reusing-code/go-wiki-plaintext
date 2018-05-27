package clean

import (
	"io/ioutil"
	"testing"
)

func TestCleanSimple(t *testing.T) {
	Clean("Test == bla")
}

func TestCleanArticle(t *testing.T) {
	content, err := ioutil.ReadFile("testdata/Notwendigkeit.txt")
	if err != nil {
		t.Fatal(err.Error())
	}

	cleaned, err := Clean(string(content))
	if err != nil {
		t.Fatal(err.Error())
	}

	err = ioutil.WriteFile("testdata/out.txt", []byte(cleaned), 0666)
	if err != nil {
		t.Fatal(err.Error())
	}
}
