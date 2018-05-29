package clean

import (
	"bytes"
	"html"
)

type cleanerState struct {
}

type elementStack struct {
	elements []element
}

type element interface {
	parseByte(b byte)
	getFinishedData() []byte
	addData([]byte)
}
type baseElement struct {
	stack       *elementStack
	data        *bytes.Buffer
	equalsCount int
	lastByte    byte
}

type headingElement struct {
	stack       *elementStack
	data        *bytes.Buffer
	level       int
	equalsCount int
}

type footerElement struct{}

type linkElement struct {
	stack    *elementStack
	data     *bytes.Buffer
	internal bool
}

type templateElement struct {
	stack    *elementStack
	lastByte byte
}

var footerHeadings = map[string]bool{
	"Notes":             true,
	"See also":          true,
	"References":        true,
	"External links":    true,
	"Suggested reading": true,
	"Siehe auch":        true,
	"Literatur":         true,
	"Weblinks":          true,
	"Einzelnachweise":   true,
	"Quellen":           true,
}

func Clean(input string) (string, error) {
	stack := newStack()
	for i := 0; i < len(input); i++ {
		stack.parseByte(input[i])
	}

	return html.UnescapeString(stack.getData()), nil
}

func newStack() *elementStack {
	stack := &elementStack{make([]element, 0)}
	stack.push(&baseElement{stack, &bytes.Buffer{}, 0, 0})
	return stack
}

func (stack *elementStack) getData() string {
	for len(stack.elements) > 1 {
		stack.pop()
	}
	return string(stack.top().getFinishedData())
}

func (stack *elementStack) top() element {
	return stack.elements[len(stack.elements)-1]
}

func (stack *elementStack) pop() {
	if len(stack.elements) > 1 {
		data := stack.top().getFinishedData()
		stack.elements = stack.elements[:len(stack.elements)-1]
		stack.top().addData(data)
	}
}

func (stack *elementStack) push(ele element) {
	stack.elements = append(stack.elements, ele)
}

func (stack *elementStack) parseByte(b byte) {
	stack.top().parseByte(b)
}

func (e *baseElement) parseByte(b byte) {
	// text formatting
	if b == '\'' {
		return
	}
	// headings
	if b == ' ' && e.equalsCount > 0 {
		e.data.Truncate(e.data.Len() - e.equalsCount)

		e.stack.push(&headingElement{e.stack, &bytes.Buffer{}, e.equalsCount, 0})

		e.equalsCount = 0
		return
	}

	if b == '=' {
		e.equalsCount++
	} else {
		e.equalsCount = 0
	}

	// links
	if e.lastByte == '[' {
		internal := false
		if b == '[' {
			internal = true
		}
		e.data.Truncate(e.data.Len() - 1)
		e.stack.push(&linkElement{e.stack, &bytes.Buffer{}, internal})
		e.lastByte = 0
		return
	}

	// templates
	if b == '{' && e.lastByte == '{' {
		e.data.Truncate(e.data.Len() - 1)
		e.stack.push(&templateElement{e.stack, 0})
		e.lastByte = 0
		return
	}

	// lists
	if e.lastByte == '#' || e.lastByte == '*' {
		e.data.Truncate(e.data.Len() - 1)
		if b != e.lastByte {
			e.lastByte = 0
			return
		}
	}

	e.lastByte = b
	e.data.WriteByte(b)
}

func (e *baseElement) getFinishedData() []byte {
	return e.data.Bytes()
}

func (e *baseElement) addData(b []byte) {
	e.data.Write(b)
}

func (e *headingElement) parseByte(b byte) {
	if b == '=' {
		if e.equalsCount == 0 {
			l := e.data.Len()
			if e.data.Bytes()[l-1] == ' ' {
				e.data.Truncate(l - 1)
			}
		}
		e.equalsCount++
		if e.equalsCount == e.level {
			e.stack.pop()
			return
		}
	} else {
		e.equalsCount = 0
		e.data.WriteByte(b)
	}

}

func (e *headingElement) getFinishedData() []byte {
	if _, ok := footerHeadings[e.data.String()]; ok {
		l := len(e.stack.elements)
		e.stack.elements = append(e.stack.elements[:l-1], &footerElement{}, e)
		return []byte{}
	} else {
		return e.data.Bytes()
	}
}

func (e *headingElement) addData(b []byte) {
	e.data.Write(b)
}

func (e *footerElement) parseByte(b byte) {
}

func (e *footerElement) getFinishedData() []byte {
	return []byte{}
}

func (e *footerElement) addData(b []byte) {
}

func (e *linkElement) parseByte(b byte) {
	if e.internal && b == '|' {
		e.data.Reset()
		return
	}
	if !e.internal && b == ' ' {
		e.data.Reset()
		return
	}
	if b == ']' && e.data.Len() > 0 {
		if e.data.Bytes()[e.data.Len()-1] == ']' {
			e.data.Truncate(e.data.Len() - 1)
			e.stack.pop()
			return
		}
		if !e.internal {
			e.stack.pop()
			return
		}
	}
	e.data.WriteByte(b)
}

func (e *linkElement) getFinishedData() []byte {
	return e.data.Bytes()
}

func (e *linkElement) addData(b []byte) {
	e.data.Write(b)
}

func (e *templateElement) parseByte(b byte) {
	if b == '}' && e.lastByte == '}' {
		e.stack.pop()
		return
	}
	e.lastByte = b
}

func (e *templateElement) getFinishedData() []byte {
	return []byte{}
}

func (e *templateElement) addData(b []byte) {
}
