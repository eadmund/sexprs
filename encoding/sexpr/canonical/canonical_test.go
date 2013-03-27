package canonical

import (
	"bytes"
	"testing"
)

func TestAtomToString(t *testing.T) {
	atom := Atom{Value: []byte("This is a test")}
	b := atom.ToCanonical()
	if !bytes.Equal(b, []byte("14:This is a test")) {
		t.Fail()
	}
}

func TestSlice(t *testing.T) {
	slice := []Sexpr{Atom{Value:[]byte("Foo")}, 
		Atom{Value:[]byte("bar")}}
	_ = slice
}

func TestList(t *testing.T) {
	var a Atom
	a = Atom{Value:[]byte("This is a test")}
	l := List{a}
	if l == nil {
		t.Fail()
	}
	t.Log(string(l.ToCanonical()))
}
