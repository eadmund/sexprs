// Copyright 2013 Robert A. Uhl.  All rights reserved.
// Use of this source code is goverend by an MIT-style license which may
// be found in the LICENSE file.

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
	s, _, err := ReadBytes([]byte("([text]test)"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = ReadBytes([]byte("(4:test3:foo(baz))"))
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

func TestTransport(t *testing.T) {
	s1, _, err := ReadBytes([]byte("{KDM6Zm9vMzpiYXJbMzpiaW5dODpiYXogcXV1eCk=}"))
	if err != nil {
		t.Fatal(err)
	}
	s2, _, err := ReadBytes([]byte("(3:foo3:bar[3:bin]8:baz quux)"))
	if err != nil {
		t.Fatal(err)
	}
	if !s1.Equal(s2) {
		t.Fatal("Transport and non-transport-loaded S-expressions are not equal")
	}
	if s2.Base64String() != ("{KDM6Zm9vMzpiYXJbMzpiaW5dODpiYXogcXV1eCk=}") {
		t.Fatal("Transport encoding failed")
	}
	t.Log(string(s1.Pack()))
}
