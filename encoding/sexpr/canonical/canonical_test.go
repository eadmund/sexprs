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
	var l Sexpr
	l = List{a}
	if l == nil {
		t.Fail()
	}
	t.Log(string(l.ToCanonical()))
}

func TestParseEmptyList(t *testing.T) {
	l, _, err := ParseBytes([]byte("()"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(l.ToCanonical()))
}

func TestParse(t *testing.T) {
	s, _, err := ParseBytes([]byte("(test)"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.ToCanonical()))
	s, _, err = ParseBytes([]byte("(4:test3:foo(baz)"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.ToCanonical()))
	s, _, err = ParseBytes([]byte("testing"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.ToCanonical()))
	s, _, err = ParseBytes([]byte("\"testing-foo bar\""))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.ToCanonical()))
	s, _, err = ParseBytes([]byte("(\"testing-foo bar\")"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.ToCanonical()))
	s, _, err = ParseBytes([]byte("(testing-foo\" bar\")"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.ToCanonical()))
	s, _, err = ParseBytes([]byte("(#7a# bar)"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.ToCanonical()))
}
