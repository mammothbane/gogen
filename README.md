# gogen

	go get github.com/mammothbane/gogen

`gogen` is a templating generics preprocessor for Go. To be clear, this means that it does not
provide runtime generics (such as you might find in Java, C#, Rust, etc.), but instead processes
generic source files into concrete implementations for each set of types that require one. Since
the Go language does not have type parameters (other than for arrays, maps, and channels), there 
are a few workarounds and caveats, and `gogen` has to be run manually (though `go generate` makes 
this very straightforward).

## Usage

Use the `generic.Generic` type to build generic types. For instance, this declaration provides
the generic type `T`:

```go
type T generic.Generic

```

which can be used more-or-less as you would normally use a generic type. For instance, to declare
a simple generic implementation of `List`:

```go
package list

import "github.com/mammothbane/gogen/generics"

type T generic.Generic

type List struct {
    data T
    next *List
}
```

We can now use `gogen` to generate a concrete implementation of `List` for any type we want:

	gogen -o intlist github.com/path/to/list T=int

which emits:

```go
package intlist
		
import "github.com/mammothbane/gogen/generic"
	
type _ generic.Generic
	
type List struct {
	data int
	next *List
}
```

as a subpackage `intlist` in the current directory. By default, this subpackage includes a `.gitignore` that
prevents it from being committed to source control, intended for use with `go generate`. To use `go generate`,
include a comment in any source file in your package as follows:

	//go:generate gogen -o intlist github.com/path/to/list T=int

`go generate ./...` will now generate the `intlist` for you. If this behavior is not desired, the `--no-gi` flag 
can be passed to `gogen`.

## Caveats

### Imports

`gogen` only supports Go builtin types at the moment; i.e. this is not possible:

	gogen -o myimpl github.com/path/to/pkg T=github.com/path/to/my/type:Type

### Zeroes and Equality

Since the underlying type of `generic.Generic` is `interface{}` and `gogen` does not make any assumptions 
when it makes its first typecheck, the following may appear to be safe because `nil` is a valid value for `interface{}`:

```go
func fn(in T) {
	if in == nil {
		// ...
	}
}
```

However, it will fail to type-check for `T=int` or any other value type. To handle this, something like the following
would work:

```go
func fn(in T) {
	var zero T
	if in == zero {
		// ...
	}
}
```

But this fails for `T=[]int`,`T=map[int]string`, etc. The more this functionality can be factored out, the better, 
but `reflect.Equal` and `reflect.DeepEqual` are also available if needed. In his `gengen` overview, joeshaw offers a nice
solution using `Equaler` interfaces, but `gogen` does not support this pattern, because `generic.Generic` (by definition)
does not implement any interfaces, and `gogen` will therefore fail on its first typecheck (before substitution). 
Support for something of this nature may be added in the future.

## Credits

This tool started out as a fork of [joeshaw's `gengen`](https://github.com/joeshaw/gengen) with
a few tweaks and fixes and ended up as a complete rewrite. Rather than manual AST loading and 
walking, `gogen` uses the [`loader` package](https://godoc.org/golang.org/x/tools/go/loader) 
and [`ast.Walk`](https://golang.org/pkg/go/ast/#Walk) to substitute and type-check references 
to generic types.

The idea for the 

```go
	type T generic.Generic
```

pattern came from [this issue](https://github.com/joeshaw/gengen/issues/2) on `gengen`.
