// Copyright 2013 Robert A. Uhl.  All rights reserved.
// Use of this source code is governed by an MIT-style license which may
// be found in the LICENSE file.

// Package sexprs implements Ron Rivest's canonical S-expressions
// (c.f. http://people.csail.mit.edu/rivest/Sexp.txt or
// rivest-draft.txt in this package) in Go.  I'm indebted to Inferno's
// sexprs(2), whose API I first accidentally, and then deliberately,
// mimicked.  I've copied much of its style, only making it more
// Go-like.
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
// The S-expression (foo bar [bin]"baz quux") is canonically:
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
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	"github.com/pkg/errors"
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
	tokenChar        = append(alpha, append(decimalDigit, simplePunc...)...)
	base64Encoding   = base64.StdEncoding
	stringChar       = append(tokenChar, append(hexadecimalDigit, []byte("\"|#")...)...)
	stringEncChar    = append(stringChar, []byte("\b\t\v\n\f\r\"'\\ ")...)
)

// Sexp is the interface implemented by both lists and atoms.
type Sexp interface {
	// String returns an advanced representation of the object, with
	// no line breaks.
	String() string

	// StringBuffer writes an advanced representation of the
	// object to a buffer, with no line breaks.
	StringBuffer(*bytes.Buffer)

	// Base64String returns a transport-encoded rendering of the
	// S-expression.
	Base64String() string

	// Pack returns the canonical representation of the object.  It
	// will always return the same sequence of bytes for the same
	// object.
	Pack() []byte

	// PackBuffer writes the canonical representation of the
	// object to a buffer.  It will always write the same sequence
	// of bytes for the same object.
	PackBuffer(*bytes.Buffer)

	// PackedLen returns the size in bytes of the canonical
	// representation.
	PackedLen() int

	// Equal will return true if its receiver and argument are
	// identical.
	Equal(b Sexp) bool
}

// A List is a slice of Lists and Atoms.
type List []Sexp

// An Atom is a byte sequence, together with an optional display hint,
// also a byte sequence.  The display hint could potentially be used
// to inform a user agent that a value should be interpreted as, e.g.,
// a JPEG.  One might use MIME types as display hints,
// e.g. [text/plain]"This is plain text" (although this particular
// case is somewhat unlikely, since most user agents are likely to
// default to plain text).  Note that there are no semantics attached
// to the value; it could be a UTF-8 string, a little-endian integer,
// a big-endian integer or any other byte sequence.
type Atom struct {
	DisplayHint []byte
	Value       []byte
}

// Pack returns the canonical form of an Atom: a decimal indicating
// its length in bytes, a colon, and then the bytes.  If there is a
// display hint, it is prepended within square brackets.  E.g. "foo"
// is packed as "3:foo" and "bar" with a display hint of "text/plain"
// is packed as "[10:text/plain]3:bar".
func (a Atom) Pack() []byte {
	buf := bytes.NewBuffer(nil)
	a.PackBuffer(buf)
	return buf.Bytes()
}

// PackBuffer implements Sexp.
func (a Atom) PackBuffer(buf *bytes.Buffer) {
	if a.DisplayHint != nil && len(a.DisplayHint) > 0 {
		buf.WriteString("[" + strconv.Itoa(len(a.DisplayHint)) + ":")
		buf.Write(a.DisplayHint)
		buf.WriteString("]")
	}
	buf.WriteString(strconv.Itoa(len(a.Value)) + ":")
	buf.Write(a.Value)
}

// PackedLen Implements Sexp.
func (a Atom) PackedLen() (size int) {
	if a.DisplayHint != nil && len(a.DisplayHint) > 0 {
		size += 3                                     // [:]
		size += len(strconv.Itoa(len(a.DisplayHint))) // decimal length
		size += len(a.DisplayHint)
	}
	size += len(strconv.Itoa(len(a.DisplayHint)))
	size++ // :
	return size + len(a.Value)
}

func (a Atom) String() string {
	buf := bytes.NewBuffer(nil)
	a.StringBuffer(buf)
	return buf.String()
}

const (
	tokenEnc = iota
	quotedEnc
	base64Enc
)

// write a string in a legible encoding to buf
func writeString(buf *bytes.Buffer, a []byte) {
	// test to see what sort of encoding is best to use
	encoding := tokenEnc
	acc := make([]byte, len(a))
	for i, c := range a {
		acc[i] = c
		switch {
		case bytes.IndexByte(tokenChar, c) > -1:
			continue
		case (encoding == tokenEnc) && bytes.IndexByte(stringEncChar, c) > -1:
			encoding = quotedEnc
			strAcc := make([]byte, i, len(a))
			copy(strAcc, acc)
			for j := i; j < len(a); j++ {
				c := a[j]
				if bytes.IndexByte(stringEncChar, c) < 0 {
					encoding = base64Enc
					break
				}
				switch c {
				case '\b':
					acc = append(strAcc, []byte("\\b")...)
				case '\t':
					strAcc = append(strAcc, []byte("\\t")...)
				case '\v':
					strAcc = append(strAcc, []byte("\\v")...)
				case '\n':
					strAcc = append(strAcc, []byte("\\n")...)
				case '\f':
					strAcc = append(strAcc, []byte("\\f")...)
				case '"':
					strAcc = append(strAcc, []byte("\\\"")...)
				case '\'':
					strAcc = append(strAcc, []byte("'")...)
				case '\\':
					strAcc = append(strAcc, []byte("\\\\")...)
				case '\r':
					strAcc = append(strAcc, []byte("\\r")...)
				default:
					strAcc = append(strAcc, c)
				}
			}
			if encoding == quotedEnc {
				buf.WriteString("\"")
				buf.Write(strAcc)
				buf.WriteString("\"")
				return
			}
		default:
			encoding = base64Enc
		}
	}
	switch encoding {
	case base64Enc:
		buf.WriteString("|" + base64Encoding.EncodeToString(acc) + "|")
	case tokenEnc:
		buf.Write(acc)
	default:
		panic("Encoding is neither base64 nor token")
	}

}

// StringBuffer implement Sexp.
func (a Atom) StringBuffer(buf *bytes.Buffer) {
	if a.DisplayHint != nil && len(a.DisplayHint) > 0 {
		buf.WriteString("[")
		writeString(buf, a.DisplayHint)
		buf.WriteString("]")
	}
	if len(a.Value) == 0 {
		buf.WriteString("\"\"")
	} else {
		writeString(buf, a.Value)
	}
}

// Base64String implements Sexp.
func (a Atom) Base64String() (s string) {
	return "{" + base64Encoding.EncodeToString(a.Pack()) + "}"
}

// Equal implements Sexp.
func (a Atom) Equal(b Sexp) bool {
	if b == nil {
		return false
	}
	switch b := b.(type) {
	case Atom:
		return bytes.Equal(a.DisplayHint, b.DisplayHint) && bytes.Equal(a.Value, b.Value)
	default:
		return false
	}
}

// Pack returns each component of List l within parentheses,
// e.g. "(foo)" would pack as "(3:foo)".
func (l List) Pack() []byte {
	buf := bytes.NewBuffer(nil)
	l.PackBuffer(buf)
	return buf.Bytes()
}

// PackBuffer implements Sexp.
func (l List) PackBuffer(buf *bytes.Buffer) {
	buf.WriteString("(")
	for _, datum := range l {
		datum.PackBuffer(buf)
	}
	buf.WriteString(")")
}

// Base64String implements Sexp.
func (l List) Base64String() string {
	return "{" + base64Encoding.EncodeToString(l.Pack()) + "}"
}

func (l List) String() string {
	buf := bytes.NewBuffer(nil)
	l.StringBuffer(buf)
	return buf.String()
}

// StringBuffer implements Sexp.
func (l List) StringBuffer(buf *bytes.Buffer) {
	buf.WriteString("(")
	for i, datum := range l {
		datum.StringBuffer(buf)
		if i < len(l)-1 {
			buf.WriteString(" ")
		}
	}
	buf.WriteString(")")
}

// Equal implements Sexp.
func (l List) Equal(b Sexp) bool {
	if l == nil && b == nil {
		return true
	}
	if b == nil {
		return false
	}
	switch b := b.(type) {
	case List:
		if len(l) != len(b) {
			return false
		}
		for i := range l {
			if !l[i].Equal(b[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// PackedLen implements Sexp.
func (l List) PackedLen() (size int) {
	size = 2 // ()
	for _, element := range l {
		size += element.PackedLen()
	}
	return size
}

// Parse returns the first S-expression in byte string s, the unparsed
// rest of s and any error encountered
func Parse(s []byte) (sexpr Sexp, rest []byte, err error) {
	//return parseSexp(bytes)
	r := bufio.NewReader(bytes.NewReader(s))
	sexpr, err = Read(r)
	if err != nil && err != io.EOF {
		return nil, nil, err
	}
	rest, err = ioutil.ReadAll(r)
	// don't confuse calling code with EOFs
	if err == io.EOF {
		err = nil
	}
	return sexpr, rest, err
}

// IsList returns true if its argument is a List.
func IsList(s Sexp) bool {
	s, ok := s.(List)
	return ok
}

// Read a single S-expression from buffered IO r, returning any error
// encountered.  May return io.EOF if at end of r; may return a valid
// S-expression and io.EOF if the EOF was encountered at the end of
// parsing.
func Read(r *bufio.Reader) (s Sexp, err error) {
	c, err := r.ReadByte()
	if err != nil {
		return nil, err
	}
	switch c {
	case '{':
		var (
			enc []byte
			n   int
		)
		if enc, err = r.ReadBytes('}'); err != nil {
			return nil, errors.Wrap(err, "couldn't read to end of transport-encoded S-expression")
		}
		acc := make([]byte, 0, len(enc)-1)
		for _, c := range enc[:len(enc)-1] {
			if bytes.IndexByte(whitespaceChar, c) == -1 {
				acc = append(acc, c)
			}
		}
		str := make([]byte, base64.StdEncoding.DecodedLen(len(acc)))
		if n, err = base64.StdEncoding.Decode(str, acc); err != nil {
			return nil, err
		}
		s, err = Read(bufio.NewReader(bytes.NewReader(str[:n])))
		if err == nil || err == io.EOF {
			return s, nil
		}
		return nil, errors.Wrap(err, "couldn't read decoded transport-encoded S-expression")
	case '(':
		l := List{}
		// skip whitespace
		for {
			var c byte
			c, err = r.ReadByte()
			if err != nil && err != io.EOF {
				return nil, errors.Wrap(err, "couldn't read next byte of list")
			}
			switch {
			case c == ')':
				return l, err
			case bytes.IndexByte(whitespaceChar, c) == -1:
				if err = r.UnreadByte(); err != nil {
					return nil, errors.Wrap(err, "couldn't unread byte")
				}
				var element Sexp
				if element, err = Read(r); err != nil {
					return nil, err
				}
				l = append(l, element)
			}
			if err != nil {
				return nil, err
			}
		}
	default:
		return readString(r, c)
	}
}

func readString(r *bufio.Reader, first byte) (s Sexp, err error) {
	var displayHint []byte
	if first == '[' {
		c, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		displayHint, err = readSimpleString(r, c)
		if err != nil {
			return nil, err
		}
		c, err = r.ReadByte()
		if err != nil {
			return nil, err
		}
		if c != ']' {
			return nil, fmt.Errorf("']' expected to end display hint; %c found", c)
		}
		first, _ = r.ReadByte() // let error be caught by readSimpleString
	}
	str, err := readSimpleString(r, first)
	return Atom{Value: str, DisplayHint: displayHint}, err
}

func readSimpleString(r *bufio.Reader, first byte) (b []byte, err error) {
	switch {
	case bytes.IndexByte(decimalDigit, first) > -1:
		return readLengthDelimited(r, first)
	case first == '#':
		return readHex(r)

	case first == '|':
		return readBase64(r)
	case first == '"':
		return readQuotedString(r, -1)
	case bytes.IndexByte(tokenChar, first) > -1:
		b = append(b, first)
		for {
			var c byte
			c, err = r.ReadByte()
			if bytes.IndexByte(tokenChar, c) == -1 {
				if err = r.UnreadByte(); err != nil {
					return nil, err
				}
				return b, err
			}
			b = append(b, c)
			if err != nil {
				return nil, err
			}
		}
	}
	return nil, errors.New("can't readSimpleString")
}

func readLengthDelimited(r *bufio.Reader, first byte) (b []byte, err error) {
	var length int64
	acc := make([]byte, 1)
	acc[0] = first
	for {
		c, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		switch {
		case bytes.IndexByte(decimalDigit, c) > -1:
			acc = append(acc, c)
		case c == ':':
			var n int
			if length, err = strconv.ParseInt(string(acc), 10, 32); err != nil {
				return nil, err
			}
			acc = make([]byte, 0, length)
			buf := make([]byte, length)
			for n, err = r.Read(buf); int64(len(acc)) < length; n, err = r.Read(buf[:length-int64(len(acc))]) {
				acc = append(acc, buf[:n]...)
				if err != nil {
					return acc, err
				}
			}
			return acc, nil
		case c == '#':
			if length, err = strconv.ParseInt(string(acc), 10, 32); err != nil {
				return nil, err
			}
			if b, err = readHex(r); err != nil {
				return nil, errors.Wrap(err, "couldn't read length-delimited bytes")
			}
			if len(b) != int(length) {
				return nil, errors.Errorf("expected %d bytes; got %d", length, len(b))
			}
			return b, err
		case c == '|':
			length, err := strconv.ParseInt(string(acc), 10, 32)
			if err != nil {
				return nil, err
			}
			if b, err = readBase64(r); err != nil {
				return nil, errors.Wrap(err, "couldn't read Base64-encoded bytes")
			}
			if len(b) != int(length) {
				return nil, errors.Errorf("expected %d bytes; got %d", length, len(b))
			}
			return b, err
		default:
			return nil, errors.Errorf("expected integer; found %c", c)
		}
	}
}

func readHex(r *bufio.Reader) (b []byte, err error) {
	var n int
	if b, err = r.ReadBytes('#'); err != nil {
		return nil, errors.Wrap(err, "couldn't read to #")
	}
	acc := make([]byte, 0, len(b)-1)
	for _, c := range b[:len(b)-1] {
		if bytes.IndexByte(whitespaceChar, c) == -1 {
			acc = append(acc, c)
		}
	}
	b = make([]byte, hex.DecodedLen(len(acc)))
	n, err = hex.Decode(b, acc)
	return b[:n], err
}

func readBase64(r *bufio.Reader) (b []byte, err error) {
	var n int
	if b, err = r.ReadBytes('|'); err != nil {
		return nil, errors.Wrap(err, "couldn't read to |")
	}
	acc := make([]byte, 0, len(b)-1)
	for _, c := range b[:len(b)-1] {
		if bytes.IndexByte(whitespaceChar, c) == -1 {
			acc = append(acc, c)
		}
	}
	b = make([]byte, base64.StdEncoding.DecodedLen(len(acc)))
	n, err = base64.StdEncoding.Decode(b, acc)
	return b[:n], err
}

type quoteState int

const (
	inQuote quoteState = iota
	inEscape
	inNewlineEscape
	inReturnEscape
	inHex1
	inHex2
	inOctal2
	inOctal3
)

func readQuotedString(r *bufio.Reader, length int) (s []byte, err error) {
	var acc, escape []byte
	if length >= 0 {
		acc = make([]byte, 0, length)
	} else {
		acc = make([]byte, 0)
	}
	escape = make([]byte, 3)
	state := inQuote
	for c, err := r.ReadByte(); err == nil; c, err = r.ReadByte() {
		switch state {
		case inQuote:
			switch c {
			case '"':
				if length > 0 && len(acc) != length {
					return nil, fmt.Errorf("Length mismatch")
				}
				return acc, err
			case '\\':
				state = inEscape
			default:
				acc = append(acc, c)
			}
		case inEscape:
			switch c {
			case byte('b'):
				acc = append(acc, '\b')
				state = inQuote
			case byte('t'):
				acc = append(acc, '\t')
				state = inQuote
			case byte('v'):
				acc = append(acc, '\v')
				state = inQuote
			case byte('n'):
				acc = append(acc, '\n')
				state = inQuote
			case byte('f'):
				acc = append(acc, '\f')
				state = inQuote
			case byte('r'):
				acc = append(acc, '\r')
				state = inQuote
			case byte('"'):
				acc = append(acc, '"')
				state = inQuote
			case byte('\''):
				acc = append(acc, '\'')
				state = inQuote
			case byte('\\'):
				acc = append(acc, '\\')
				state = inQuote
			case byte('\n'):
				state = inNewlineEscape
			case '\r':
				state = inReturnEscape
			case byte('x'):
				state = inHex1
			default:
				if bytes.IndexByte(octalDigit, c) > -1 {
					state = inOctal2
					escape[0] = c
				} else {
					return nil, fmt.Errorf("Unrecognised escape character %c", rune(c))
				}
				state = inQuote
			}
		case inNewlineEscape:
			switch c {
			case '\r':
				// pass
			case '"':
				if length > 0 && len(acc) != length {
					return nil, fmt.Errorf("Length mismatch")
				}
				return acc, nil
			default:
				acc = append(acc, c)
			}
			state = inQuote
		case inReturnEscape:
			switch c {
			case '\n':
				// pass
			case '"':
				if length > 0 && len(acc) != length {
					return nil, fmt.Errorf("Length mismatch")
				}
				return acc, nil
			default:
				acc = append(acc, c)
			}
			state = inQuote
		case inHex1:
			if bytes.IndexByte(hexadecimalDigit, c) > -1 {
				state = inHex2
				escape[0] = c
			} else {
				return nil, fmt.Errorf("Expected hexadecimal digit; got %c", c)
			}
		case inHex2:
			if bytes.IndexByte(hexadecimalDigit, c) > -1 {
				state = inQuote
				escape[2] = c
				num, err := strconv.ParseInt(string(escape[:2]), 16, 8)
				if err != nil {
					return nil, err
				}
				acc = append(acc, byte(num))
			} else {
				return nil, fmt.Errorf("Expected hexadecimal digit; got %c", c)
			}
		case inOctal2:
			if bytes.IndexByte(octalDigit, c) > -1 {
				state = inOctal3
				escape[1] = c
			} else {
				return nil, fmt.Errorf("Expected octal digit; got %c", c)
			}
		case inOctal3:
			if bytes.IndexByte(octalDigit, c) > -1 {
				state = inQuote
				escape[2] = c
				num, err := strconv.ParseInt(string(escape[:2]), 8, 8)
				if err != nil {
					return nil, err
				}
				acc = append(acc, byte(num))
			} else {
				return nil, fmt.Errorf("Expected octal digit; got %c", c)
			}
		}
	}
	return nil, fmt.Errorf("Unterminated string")
}
