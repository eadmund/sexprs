# sexprs
--
    import "github.com/eadmund/sexprs"

Package sexprs implements Ron Rivest's canonical S-expressions
<URL:http://people.csail.mit.edu/rivest/Sexp.txt> in Go.  I'm
indebted to Inferno's sexprs(2), whose API I first accidentally,
and then deliberately, mimicked.  I've copied much of its style,
only making it more Go-like.

Canonical S-expressions are a compact, easy-to-parse, ordered,
hashable data representation ideal for cryptographic operations.
They are simpler and more compact than either JSON or XML.

An S-expression is composed of lists and atoms.  An atom is a string
of bytes, with an optional display hint, also a byte string.  A list
can contain zero or more atoms or lists.

There are two representations of an S-expression: the canonical
representation is a byte-oriented, packed representation, while the
advanced representation is string-oriented and more traditional in
appearance.

The S-expression ("foo" "bar" ["bin"]"baz quux") is canonically:
   (3:foo3:bar[3:bin]8:quux)

Among the valid advanced representations are:
   (foo 3:bar [bin]"baz quux")
and:
   ("foo" #626172# [3:bin]|YmF6IHF1dXg=|)

There is also a transport encoding (intended for use in 7-bit transport
modes), delimited with {}:
   {KDM6Zm9vMzpiYXJbMzpiaW5dODpiYXogcXV1eCk=}

## Usage

#### func  IsList

```go
func IsList(s Sexp) bool
```

#### type Atom

```go
type Atom struct {
	DisplayHint []byte
	Value       []byte
}
```


#### func (Atom) Base64String

```go
func (a Atom) Base64String() (s string)
```

#### func (Atom) Equal

```go
func (a Atom) Equal(b Sexp) bool
```

#### func (Atom) Pack

```go
func (a Atom) Pack() []byte
```

#### func (Atom) PackedLen

```go
func (a Atom) PackedLen() (size int)
```

#### func (Atom) String

```go
func (a Atom) String() string
```

#### type List

```go
type List []Sexp
```


#### func (List) Base64String

```go
func (l List) Base64String() string
```

#### func (List) Equal

```go
func (a List) Equal(b Sexp) bool
```

#### func (List) Pack

```go
func (l List) Pack() []byte
```

#### func (List) PackedLen

```go
func (l List) PackedLen() (size int)
```

#### func (List) String

```go
func (l List) String() string
```

#### type Sexp

```go
type Sexp interface {
	// String returns an advanced representation of the object, with
	// no line breaks.
	String() string

	// Base64String returns a transport-encoded rendering of the
	// S-expression
	Base64String() string

	// Pack returns the canonical representation of the object.  It
	// will always return the same sequence of bytes for the same
	// object.
	Pack() []byte

	// PackedLen returns the size in bytes of the canonical
	// representation.
	PackedLen() int

	// Equal will return true if its receiver and argument are
	// identical.
	Equal(b Sexp) bool
	// contains filtered or unexported methods
}
```

Sexp is the interface implemented by both lists and atoms.

#### func  Parse

```go
func Parse(s []byte) (sexpr Sexp, rest []byte, err error)
```

#### func  Read

```go
func Read(r *bufio.Reader) (s Sexp, err error)
```
