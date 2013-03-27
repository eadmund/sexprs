package canonical

import (
	"bytes"
	"strconv"
)

type Sexpr interface {
	ToAdvanced() string
	toAdvanced(*bytes.Buffer)
	ToCanonical() []byte
	toCanonical(*bytes.Buffer)
}

type List []Sexpr

type Atom struct {
	DisplayHint []byte
	Value       []byte
}

func (a Atom) ToCanonical() []byte {
	buf := bytes.NewBuffer(nil)
	a.toCanonical(buf)
	return buf.Bytes()
}

func (a Atom) toCanonical(buf *bytes.Buffer) {
	if a.DisplayHint != nil && len(a.DisplayHint) > 0 {
		buf.WriteString("[" + strconv.Itoa(len(a.DisplayHint)) + ":")
		buf.Write(a.DisplayHint)
		buf.WriteString("]")
	}
	buf.WriteString(strconv.Itoa(len(a.Value)) + ":")
	buf.Write(a.Value)
}

func (a Atom) ToAdvanced() string {
	buf := bytes.NewBuffer(nil)
	a.toAdvanced(buf)
	return buf.String()
}

func (a Atom) toAdvanced(buf *bytes.Buffer) {
	buf.WriteString(strconv.Itoa(len(a.Value)) + ":")
	buf.Write(a.Value)
}

// func Cons(car Sexpr, cdr Sexpr) Cons {
// 	return Cons{car, cdr}
// }

func (l List) ToCanonical() []byte {
	buf := bytes.NewBuffer(nil)
	l.toCanonical(buf)
	return buf.Bytes()
}

func (l List) toCanonical(buf *bytes.Buffer) {
	buf.WriteString("(")
	for _, datum := range l {
		datum.toCanonical(buf)
	}
	buf.WriteString(")")
}

func (l List) ToAdvanced() string {
	buf := bytes.NewBuffer(nil)
	l.toAdvanced(buf)
	return buf.String()
}

func (l List) toAdvanced(buf *bytes.Buffer) {
	buf.WriteString("(")
	for _, datum := range l {
		datum.toAdvanced(buf)
	}
	buf.WriteString(")")
}

func ParseBytes(bytes []byte) []Sexpr {
	return nil
}
