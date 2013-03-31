# sexprs
--
    import "github.com/eadmund/sexprs"

Package sexprs implements Ron Rivest's canonical s-expression
<URL:http://people.csail.mit.edu/rivest/Sexp.txt> in Go.

## Usage

#### type Atom

```go
type Atom struct {
	DisplayHint []byte
	Value       []byte
}
```


#### func (Atom) Pack

```go
func (a Atom) Pack() []byte
```

#### func (Atom) String

```go
func (a Atom) String() string
```

#### type List

```go
type List []Sexp
```


#### func (List) Pack

```go
func (l List) Pack() []byte
```

#### func (List) String

```go
func (l List) String() string
```

#### type Sexp

```go
type Sexp interface {
	String() string

	Pack() []byte
	// contains filtered or unexported methods
}
```


#### func  ReadBytes

```go
func ReadBytes(bytes []byte) (sexpr Sexp, rest []byte, err error)
```
