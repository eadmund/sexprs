// Copyright 2013 Robert A. Uhl.  All rights reserved.
// Use of this source code is goverend by an MIT-style license which may
// be found in the LICENSE file.

package sexprs

import (
	"bufio"
	"bytes"
	"io"
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
	slice := []Sexp{Atom{Value: []byte("Foo")},
		Atom{Value: []byte("bar")}}
	_ = slice
}

func TestList(t *testing.T) {
	var a Atom
	a = Atom{Value: []byte("This is a test")}
	var l Sexp
	l = List{a}
	if l == nil {
		t.Fail()
	}
	t.Log(string(l.Pack()))
}

func TestParseEmptyList(t *testing.T) {
	l, _, err := Parse([]byte("()"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(l.Pack()))
}


func TestError(t *testing.T) {
	confirmError := func(a string) {
		_, _, err := Parse([]byte(a))
		if err == nil {
			t.Fatalf("Parsing %v should have produced an error", a)
		}
	}
	confirmError("((a)")
}

func TestParse(t *testing.T) {
	s, _, err := Parse([]byte("([text]test)"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = Parse([]byte("(4:test3:foo(baz))"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = Parse([]byte("testing"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = Parse([]byte("\"testing-foo bar\""))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = Parse([]byte("(\"testing-foo bar\")"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = Parse([]byte("(testing-foo\" bar\")"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(s.Pack()))
	s, _, err = Parse([]byte("([foo/bar]#7a # [\"quux beam\"]bar ([jim]|Zm9vYmFy YmF6|)\"foo bar\\r\"{Zm9vYmFyYmF6})"))
	if err != nil {
		t.Fatal(err)
	}
	if s == nil {
		t.Fatal("Not parsed")
	}
	t.Log(string(s.Pack()))
	t.Log(s.String())
}

func TestTransport(t *testing.T) {
	s1, _, err := Parse([]byte("{KDM6Zm9vMzpiYXJbMzpiaW5dODpiYXogcXV1eCk=}"))
	if err != nil {
		t.Fatal(err)
	}
	if s1 == nil {
		t.Fatal("List is nil")
	}
	s2, _, err := Parse([]byte("(3:foo3:bar[3:bin]8:baz quux)"))
	if err != nil {
		t.Fatal(err)
	}
	if s2 == nil {
		t.Fatal("List is nil")
	}
	if !s1.Equal(s2) {
		t.Fatal("Transport and non-transport-loaded S-expressions are not equal")
	}
	if s2.Base64String() != ("{KDM6Zm9vMzpiYXJbMzpiaW5dODpiYXogcXV1eCk=}") {
		t.Fatal("Transport encoding failed")
	}
	t.Log(string(s1.Pack()))
}

func TestIsList(t *testing.T) {
	s, _, err := Parse([]byte("(abc efg-hijk )"))
	if err != nil {
		t.Fatal("Could not parse list", err)
	}
	if !IsList(s) {
		t.Fatal("List considered not-List")
	}
	s, _, err = Parse([]byte("abc"))
	if err != nil {
		t.Fatal("Could not parse atom", err)
	}
	if IsList(s) {
		t.Fatal("Atom considered List")
	}
}

func TestRead(t *testing.T) {
	s, err := Read(bufio.NewReader(bytes.NewReader([]byte("()"))))
	if err != nil {
		t.Fatal(err)
	}
	l, ok := s.(List)
	if !ok {
		t.Fatal("List expected")
	}
	if len(l) != 0 {
		t.Fatal("Zero-length list expected")
	}
	s, err = Read(bufio.NewReader(bytes.NewReader([]byte("6:foobar"))))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	a, ok := s.(Atom)
	if !ok {
		t.Fatal("Atom expected")
	}
	t.Log(string(a.Value), len(a.Value))
	s, err = Read(bufio.NewReader(bytes.NewReader([]byte("7:foobar"))))
	if err == nil {
		t.Fatal("Didn't fail on invalid bytestring")
	}
	s, err = Read(bufio.NewReader(bytes.NewReader([]byte("3#61 6 263#"))))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	a, ok = s.(Atom)
	if !ok {
		t.Fatal("Atom expected")
	}
	if !bytes.Equal(a.Value, []byte("abc")) {
		t.Fatal("Bad ", a)
	}
	s, err = Read(bufio.NewReader(bytes.NewReader([]byte("3|Y2\r\nJ h|"))))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	a, ok = s.(Atom)
	if !ok {
		t.Fatal("Atom expected")
	}
	if !bytes.Equal(a.Value, []byte("cba")) {
		t.Fatal("Bad ", a)
	}
	//t.Log(">>", string(a.Value))
	// hex without length
	s, err = Read(bufio.NewReader(bytes.NewReader([]byte("#616263#"))))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	a, ok = s.(Atom)
	if !ok {
		t.Fatal("Atom expected")
	}
	if !bytes.Equal(a.Value, []byte("abc")) {
		t.Fatal("Bad ", a)
	}
	// base64 without length
	s, err = Read(bufio.NewReader(bytes.NewReader([]byte("|Y2Jh|"))))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	a, ok = s.(Atom)
	if !ok {
		t.Fatal("Atom expected")
	}
	if !bytes.Equal(a.Value, []byte("cba")) {
		t.Fatal("Bad ", a)
	}
	// quoted string without length
	s, err = Read(bufio.NewReader(bytes.NewReader([]byte("\"Foo bar \rbaz quux\\\nquuux\""))))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	a, ok = s.(Atom)
	if !ok {
		t.Fatal("Atom expected")
	}
	if !bytes.Equal(a.Value, []byte("Foo bar \rbaz quuxquuux")) {
		t.Fatal("Bad ", a)
	}
	// escaped return
	s, err = Read(bufio.NewReader(bytes.NewReader([]byte("\"Foo bar \\\r\""))))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	a, ok = s.(Atom)
	if !ok {
		t.Fatal("Atom expected")
	}
	if !bytes.Equal(a.Value, []byte("Foo bar ")) {
		t.Fatal("Bad ", a)
	}
	// list
	s, err = Read(bufio.NewReader(bytes.NewReader([]byte("(a b)"))))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	l, ok = s.(List)
	if !ok {
		t.Fatal("List expected")
	}
	if !l.Equal(List{Atom{Value: []byte("a")}, Atom{Value: []byte("b")}}) {
		t.Fatal("Bad ", l)
	}
	// display hint
	s, err = Read(bufio.NewReader(bytes.NewReader([]byte("[abc]bar"))))
	if err != nil && err != io.EOF {
		t.Fatal(err)
	}
	if !s.Equal(Atom{DisplayHint: []byte("abc"), Value: []byte("bar")}) {
		t.Fatal("Bad s-expression", s)
	}
}
