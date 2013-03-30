package sexprs

import (
	"bytes"
	"testing"
)

func TestAtomToString(t *testing.T) {
	atom := Atom{Value: []byte("This is a test")}
	b := atom.Pack()
	if !bytes.Equal(b, []byte("14:This is a test")) {
		t.Fail()
	}
}

func TestSlice(t *testing.T) {
	slice := []Sexp{Atom{Value:[]byte("Foo")}, 
		Atom{Value:[]byte("bar")}}
	_ = slice
}

func TestList(t *testing.T) {
	var a Atom
	a = Atom{Value:[]byte("This is a test")}
	var l Sexp
	l = List{a}
	if l == nil {
		t.Fail()
	}
	t.Log(string(l.Pack()))
}

func TestParseEmptyList(t *testing.T) {
	l, _, err := ReadBytes([]byte("()"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(l.Pack()))
}

func TestParse(t *testing.T) {
	s, _, err := ReadBytes([]byte("(test)"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = ReadBytes([]byte("(4:test3:foo(baz)"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = ReadBytes([]byte("testing"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = ReadBytes([]byte("\"testing-foo bar\""))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = ReadBytes([]byte("(\"testing-foo bar\")"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = ReadBytes([]byte("(testing-foo\" bar\")"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = ReadBytes([]byte("(#7a# bar)"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
}
