# gogen [![Build Status](https://travis-ci.org/mammothbane/gogen.svg?branch=master)](https://travis-ci.org/mammothbane/gogen) [![Go report](http://goreportcard.com/badge/mammothbane/gogen)](http://goreportcard.com/report/mammothbane/gogen)

	go get github.com/mammothbane/gogen

`gogen` is a templating generics preprocessor for Go. It allows you to write pseudo-generic code like this: 

```go
package list

import "github.com/mammothbane/gogen/generics”

type T generic.Generic

type List struct {
    data T
    next *List
}
```

And, as it turns out, that's valid Go! `generic.Generic` resolves to `interface{}`, which means that this snippet could be used as-is with type assertions. However, if we run `gogen`:

    gogen -o intlist github.com/path/to/list T=int

A new package `intlist` is created for us in the current directory with `list.go` as the only file:

```go
package intlist

import "github.com/mammothbane/gogen/generic"

type _ generic.Generic

type List struct {
	data int
	next *List
}
```

`intlist` can now be imported and used normally. 


## Usage
Rename `generic.Generic` to create generic types. This declaration provides the generic types `T` and `U`:

```go
type T generic.Generic
type U generic.Generic
```

You can define as many types as you want in this manner, as long as their names don't collide with any other identifiers in the package. To generate a concrete implementation of the package, run `gogen`:

	gogen -o [outpkg] [inpkg] K=V K=V ... 

Where `outpkg` is the path of the package to be created, `inpkg` is the location of the generic package in your `GOPATH` (usually of the form `github.com/username/packages...`), and the `K=V` pairs map from generic type names to concrete types.

Note that because Go has no support for type parameters (e.g. `func[U] fn() U`, as you might be used to seeing in Java, Scala, C#, or other similar languages), all generic types are declared at the **package scope**; i.e. globally. So for any given `func() T` in your package, `T` represents the same type. Of course, `T` could be an interface type, or even `interface{}`, if you desire, but using `interface{}` somewhat defeats the point of using generics in the first place.

## Go generate
As you're writing and maintaining your generic package, it may become tedious to write out your entire `gogen` command each time you make a change. It's suggested that you use `go generate` to handle this process for you. Simply add

	//go:generate gogen -o [outpkg] github.com/path/to/generic/pkg Key=Val Key=Val ... 

to any source file in your package, and whenever you run `go generate` in your package directory (or `go generate ./...` in a parent), your package will be (re)generated. It may make sense to put this in a `generate.go` file or similar with `// +build ignore`.

You may also want to add `gogen` to your build process (`go get github.com/mammothbane/gogen && go generate ./...`) rather than committing the generated files to source control. 

By default, `gogen` creates a `.gitignore` in the generated package set to ignore everything. To disable this behavior, pass the `--no-gi` flag to `gogen`.

## Inner workings
Feel free to look at `main.go` for the bulk of the implementation. At the basic level, `gogen`:

- loads and type-checks the generic package **G**
- loads and type-checks `github.com/mammothbane/gogen/generic`
- gets the type **T** of `generic.Generic` from `github.com/mammothbane/gogen/generic`
- searches the file ASTs in **G** for type specs of type **T**, eliding those that match the commandline KV pairs with `_`
- searches the file ASTs in **G** for identifiers matching the KV pairs and replaces them with concrete types
- rewrites the package names for all file ASTs in **G**
- writes out the new package, then reloads and type-checks it

## Caveats
### Imports
`gogen` only supports Go builtin types at the moment; i.e. this is not possible (but it will be soon):

	gogen -o myimpl github.com/path/to/pkg T=github.com/path/to/my/type:Type

### Zeroes and Equality
Since the underlying type of `generic.Generic` is `interface{}` and `gogen` does not make any assumptions 
when it makes its first type-check, the following may appear to be safe because `nil` is a valid value for `interface{}`:

```go
func fn(in T) {
	if in == nil {
		// ...
	}
}
```

However, it will fail to type-check for `T=int` or any other value type. To handle this, something like the following would work:

```go
func fn(in T) {
	var zero T
	if in == zero {
		// ...
	}
}
```

But this fails for `T=[]int`,`T=map[int]string`, etc. The more this functionality can be factored out, the better, but `reflect.Equal` and `reflect.DeepEqual` are also available if needed. 

In his `gengen` overview, joeshaw offers a nice solution using `Equaler` interfaces, but `gogen` does not support this pattern, because `generic.Generic` (by definition) does not implement any interfaces, and `gogen` will therefore fail on its first typecheck (before substitution). Support for something of this nature may be added in the future.

## Credits
This tool started out as a fork of [joeshaw's `gengen`](https://github.com/joeshaw/gengen) with a few tweaks and fixes and ended up as a complete rewrite. Rather than manual AST loading and walking, `gogen` uses the [`loader` package](https://godoc.org/golang.org/x/tools/go/loader) and [`ast.Walk`](https://golang.org/pkg/go/ast/#Walk) to substitute and type-check references to generic types.

The idea for the 

```go
	type T generic.Generic
```

pattern came from [this issue](https://github.com/joeshaw/gengen/issues/2) on `gengen`.
