// Copyright 2013 Robert A. Uhl.  All rights reserved.
// Use of this source code is governed by an MIT-style license which may
// be found in the LICENSE file.

// Package sexprs implements Ron Rivest's canonical S-expressions
// <URL:http://people.csail.mit.edu/rivest/Sexp.txt> in Go.  I'm
// indebted to Inferno's sexprs(2), whose API I first accidentally,
// and then deliberately, mimicked.  I've copied much of its style,
// only making it more Go-like.
//
// Canonical S-expressions are a compact, easy-to-parse, ordered,
// hashable data representation ideal for cryptographic operations.
// They are simpler and more compact than either JSON or XML.
//
// An S-expression is composed of lists and atoms.  An atom is a string
// of bytes, with an optional display hint, also a byte string.  A list
// can contain zero or more atoms or lists.
//
// There are two representations of an S-expression: the canonical
// representation is a byte-oriented, packed representation, while the
// advanced representation is string-oriented and more traditional in
// appearance.
//
// The S-expression ("foo" "bar" ["bin"]"baz quux") is canonically:
//    (3:foo3:bar[3:bin]8:quux)
//
// Among the valid advanced representations are:
//    (foo 3:bar [bin]"baz quux")
// and:
//    ("foo" #626172# [3:bin]|YmF6IHF1dXg=|)
// 
// There is also a transport encoding (intended for use in 7-bit transport
// modes), delimited with {}:
//    {KDM6Zm9vMzpiYXJbMzpiaW5dODpiYXogcXV1eCk=}
//
package sexprs

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
)

var (
	lowerCase        = []byte("abcdefghijklmnopqrstuvwxyz")
	upperCase        = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	decimalDigit     = []byte("0123456789")
	alpha            = append(lowerCase, upperCase...)
	hexadecimalDigit = append(decimalDigit, []byte("abcdefABCDEF")...)
	octalDigit       = []byte("01234567")
	simplePunc       = []byte("-./_:*+=")
	whitespaceChar   = []byte(" \t\r\n")
	base64Char       = append(alpha, append(decimalDigit, []byte("+/=")...)...)
	tokenChar        = append(alpha, append(decimalDigit, simplePunc...)...)
	base64Encoding    = base64.StdEncoding
	stringChar       = append(tokenChar, append(hexadecimalDigit, []byte("\"|#")...)...)
)

// Sexp is the interface implemented by any object with an S-expression
// representation.  It's not really intended to be implemented outside
// of sexprs, although it's certainly possible.
type Sexp interface {
	// String returns an advanced representation of the object, with
	// no line breaks.
	String() string
	string(*bytes.Buffer)

	// Base64String returns a transport-encoded rendering of the
	// S-expression
	Base64String() string

	// Pack returns the canonical representation of the object.  It
	// will always return the same sequence of bytes for the same
	// object.
	Pack() []byte
	pack(*bytes.Buffer)

	// PackedLen returns the size in bytes of the canonical
	// representation.
	PackedLen() int

	// Equal will return true if its receiver and argument are
	// identical.
	Equal(b Sexp) bool
}

type List []Sexp

type Atom struct {
	DisplayHint []byte
	Value       []byte
}

func (a Atom) Pack() []byte {
	buf := bytes.NewBuffer(nil)
	a.pack(buf)
	return buf.Bytes()
}

func (a Atom) pack(buf *bytes.Buffer) {
	if a.DisplayHint != nil && len(a.DisplayHint) > 0 {
		buf.WriteString("[" + strconv.Itoa(len(a.DisplayHint)) + ":")
		buf.Write(a.DisplayHint)
		buf.WriteString("]")
	}
	buf.WriteString(strconv.Itoa(len(a.Value)) + ":")
	buf.Write(a.Value)
}

func (a Atom) PackedLen() (size int) {
	if a.DisplayHint != nil && len(a.DisplayHint) > 0 {
		size += 3 // [:]
		size += len(strconv.Itoa(len(a.DisplayHint))) // decimal length
		size += len(a.DisplayHint)
	}
	size += len(strconv.Itoa(len(a.DisplayHint)))
	size++ // :
	return size + len(a.Value)
}

func (a Atom) String() string {
	buf := bytes.NewBuffer(nil)
	a.string(buf)
	return buf.String()
}

func (a Atom) string(buf *bytes.Buffer) {
	buf.WriteString(strconv.Itoa(len(a.Value)) + ":")
	buf.Write(a.Value)
}

func (a Atom) Base64String() (s string) {
	return "{" + base64Encoding.EncodeToString(a.Pack()) + "}"
}

func (a Atom) Equal(b Sexp) bool {
	switch b := b.(type) {
	case Atom:
		return bytes.Equal(a.DisplayHint, b.DisplayHint) && bytes.Equal(a.Value, b.Value)
	default:
		return false
	}
	return false
}

func (l List) Pack() []byte {
	buf := bytes.NewBuffer(nil)
	l.pack(buf)
	return buf.Bytes()
}

func (l List) pack(buf *bytes.Buffer) {
	buf.WriteString("(")
	for _, datum := range l {
		datum.pack(buf)
	}
	buf.WriteString(")")
}

func (l List) Base64String() string {
	return "{" + base64Encoding.EncodeToString(l.Pack()) + "}"
}

func (l List) String() string {
	buf := bytes.NewBuffer(nil)
	l.string(buf)
	return buf.String()
}

func (l List) string(buf *bytes.Buffer) {
	buf.WriteString("(")
	for _, datum := range l {
		datum.string(buf)
	}
	buf.WriteString(")")
}

func (a List) Equal(b Sexp) bool {
	switch b := b.(type) {
	case List:
		if len(a) != len(b) {
			return false
		} else {
			for i := range a {
				if !a[i].Equal(b[i]) {
					return false
				}
			}
			return true
		}
	default:
		return false
	}
	return false
}

func (l List) PackedLen() (size int) {
	size = 2 // ()
	for _, element := range l {
		size += element.PackedLen()
	}
	return size
}

func ReadBytes(bytes []byte) (sexpr Sexp, rest []byte, err error) {
	return parseSexp(bytes)

}

func parseSexp(s []byte) (sexpr Sexp, rest []byte, err error) {
	first, rest := s[0], s[1:]
	switch {
	case first == byte('('):
		return parseList(rest)
	case first == byte('{'):
		return parseTransport(rest)
	case bytes.IndexByte(stringChar, first) > -1, first == byte('['):
		return parseAtom(s)
	default:
		return nil, rest, fmt.Errorf("Unrecognised character at start of s-expression: %c", first)
	}

	panic("Should never get here")
}

func parseList(s []byte) (l List, rest []byte, err error) {
	acc := make(List, 0)
	var sexpr Sexp
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == byte(')'):
			return acc, s[i+1:], nil
		case bytes.IndexByte(whitespaceChar, c) == -1:
			sexpr, s, err = parseSexp(s[i:])
			if err != nil {
				return nil, nil, err
			}
			i = -1
			acc = append(acc, sexpr)
		}
	}
	return nil, nil, fmt.Errorf("Expected ')' to terminate list")
}

func parseAtom(s []byte) (a Atom, rest []byte, err error) {
	first, rest := s[0], s[1:]
	var displayHint, value []byte
	if first == byte('[') {
		displayHint, s, err = parseSimpleString(rest)
		if err != nil {
			return Atom{}, rest, err
		}
		s = s[1:]
	}
	value, rest, err = parseSimpleString(s)
	if err != nil {
		return Atom{}, nil, err
	}
	return Atom{DisplayHint: displayHint, Value: value}, rest, nil
}

func parseSimpleString(s []byte) (str, rest []byte, err error) {
	length := -1
	if bytes.IndexByte(decimalDigit, s[0]) > -1 {
		var lengthString []byte
		lengthString, s, err = parseDecimal(s)
		if err != nil {
			return nil, nil, err
		}
		length, err = strconv.Atoi(string(lengthString))
		if err != nil {
			return nil, nil, err
		}
	}
	switch s[0] {
	case byte(':'):
		if length < 0 {
			return nil, nil, fmt.Errorf("Unspecified length of raw string")
		}
		return s[1 : length+1], s[length+1:], nil
	case byte('#'):
		str, rest, err = parseHexadecimal(s[1:])
	case byte('|'):
		str, rest, err = parseBase64(s[1:])
	case byte('"'):
		str, rest, err = parseQuotedString(s[1:], length)
	default:
		if bytes.IndexByte(tokenChar, s[0]) > -1 {
			var i int
			for i = 1; i < len(s) && bytes.IndexByte(tokenChar, s[i]) > -1; i++ {
			}
			str = s[0:i]
			return str, s[i:], nil
		}
		return nil, nil, fmt.Errorf("Unknown char %c parsing simple string", s[0])
	}
	if err != nil {
		return nil, nil, err
	}
	if length != -1 {
		if len(str) != length {
			return nil, nil, fmt.Errorf("Explicit length %d not equal to implicit length %d", length, len(str))
		}
		return str, s[length:], nil
	}
	return str, rest, nil
}

func parseDecimal(s []byte) (decimal, rest []byte, err error) {
	for i := range s {
		if bytes.IndexByte(decimalDigit, s[i]) < 0 {
			return s[0:i], s[i:], nil
		}
	}
	return s, rest, nil
}

func parseHexadecimal(s []byte) (str, rest []byte, err error) {
	for i := range s {
		if bytes.IndexByte(hexadecimalDigit, s[i]) < 0 {
			if s[i] != byte('#') {
				return nil, nil, fmt.Errorf("Expected # to terminate hexadecimal string; found %c", s[i])
			}
			str := make([]byte, hex.DecodedLen(i))
			length, err := hex.Decode(str, s[0:i])
			if err != nil {
				return nil, nil, err
			}
			return str[:length], s[i+1:], nil
		}
	}
	return nil, nil, fmt.Errorf("Unexpected end of hexadecimal value")
}

func parseBase64(s []byte) (decimal, rest []byte, err error) {
	for i := range s {
		if bytes.IndexByte(hexadecimalDigit, s[i]) < 0 {
			if s[i] != byte('|') {
				return nil, nil, fmt.Errorf("Expected | to terminate Base64 string")
			}
			base64 := s[0:i]
			decimal = make([]byte, base64Encoding.DecodedLen(len(base64)))
			length, err := base64Encoding.Decode(decimal, base64)
			if err != nil {
				return nil, nil, err
			}
			return base64[:length], s[i:], nil
		}
	}
	return nil, nil, fmt.Errorf("Unexpected end of Base64 value")
}

func parseQuotedString(s []byte, length int) (decimal, rest []byte, err error) {
	var acc []byte
	if length > 0 {
		acc = make([]byte, length)
	} else {
		acc = make([]byte, 0)
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case byte('"'):
			if length != -1 && len(acc) != length {
				return nil, nil, fmt.Errorf("Explicit length %d not equal to implicit length %d", length, len(acc))
			}
			return acc, s[i+1:], nil
		case '\\':
			i++
			if i == len(s) {
				return nil, nil, fmt.Errorf("Unterminated quoted string")
			}
			c = s[i]
			switch c {
			case byte('b'):
				c = byte('\b')
			case byte('t'):
				c = byte('\t')
			case byte('v'):
				c = byte('\v')
			case byte('n'):
				c = byte('\n')
			case byte('f'):
				c = byte('\f')
			case byte('r'):
				c = byte('\r')
			case byte('"'):
				c = byte('"')
			case byte('\''):
				c = byte('\'')
			case byte('\\'):
				c = byte('\\')
			case byte('\n'):
				if i+1 < len(s) && s[i+1] == byte('\r') {
					i++
				}
				continue
			case byte('\r'):
				if i+1 < len(s) && s[i+1] == byte('\n') {
					i++
				}
				continue
			case byte('x'):
				num, err := strconv.ParseInt(string(s[i+1:i+2]), 16, 8)
				if err != nil {
					return nil, nil, err
				}
				c = byte(num)
			default:
				if bytes.IndexByte(octalDigit, c) > -1 && bytes.IndexByte(octalDigit, s[i+1]) > -1 && bytes.IndexByte(octalDigit, s[i+2]) > -1 {
					num, err := strconv.ParseInt(string(s[i:i+2]), 8, 8)
					if err != nil {
						return nil, nil, err
					}
					c = byte(num)
				}
				return nil, nil, fmt.Errorf("Unrecognised escape character %c", rune(c))
			}
			fallthrough
		default:
			acc = append(acc, c)
		}
	}
	return nil, nil, fmt.Errorf("Unexpected end of quoted string")
}

func parseTransport(s []byte) (sexp Sexp, rest []byte, err error) {
	for i := range s {
		if s[i] == byte('}') {
			decoded := make([]byte, base64Encoding.DecodedLen(i))
			length, err := base64Encoding.Decode(decoded, s[0:i])
			if err != nil {
				return nil, nil, err
			}
			sexp, rest, err = parseSexp(decoded[:length])
			if len(rest) != 0 {
				return nil, nil, fmt.Errorf("Expected complete single transport-encoded S-expression")
			}
			return sexp, s[i+1:], err
		}
	}
	return nil, nil, fmt.Errorf("Expected '}' to terminate transport representation")
}
