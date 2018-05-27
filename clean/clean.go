package clean

import (
	"bytes"
	"html"
)

type cleanerState struct {
}

type elementStack struct {
	stack []element
	data  *bytes.Buffer
}

type element struct {
	elemType elementType
	data     *bytes.Buffer
}

type elementType int

const (
	HEADING_1 elementType = iota
	HEADING_2
	HEADING_3
	HEADING_4
	HEADING_5
	INTERNAL_LINK
	EXTERNAL_LINK
)

func Clean(input string) (string, error) {
	stack := &elementStack{make([]element, 0), &bytes.Buffer{}}
	for i := 0; i < len(input); i++ {
		a := input[i]
		b := '"'
		c := string(a)
		if input[i] == '=' {
			continue
		}
		stack.data.WriteByte(input[i])
		a = a + byte(b)
		c = c + c
	}

	return html.UnescapeString(stack.data.String()), nil
}
