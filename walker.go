package main

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/loader"
)

type (
	walker struct {
		mp          map[string]string
		pkgInfo     *loader.PackageInfo
		genericType *types.Type
	}

	genericWalker walker
	typeWalker    walker
	nameWalker    walker
)

func (w *genericWalker) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	if n, ok := node.(*ast.ValueSpec); ok {
		nType := w.pkgInfo.TypeOf(n.Type)
		w.genericType = &nType
		return nil
	}

	return w
}

func (w *typeWalker) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	if n, ok := node.(*ast.TypeSpec); ok {
		if _, ok := w.mp[n.Name.Name]; ok && w.pkgInfo.TypeOf(n.Type) == *w.genericType {
			n.Name.Name = "_"
		}
	}

	return w
}

func (w *nameWalker) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		return nil
	}

	if n, ok := node.(*ast.Ident); ok {
		if newName, ok := w.mp[n.Name]; ok {
			n.Name = newName
		}
	}

	return w
}
