package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"io/ioutil"
	stdlog "log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/tools/go/loader"
)

const genericsPackageName = "github.com/mammothbane/gogen/generic"

var (
	quiet   bool
	verbose bool
)

func main() {
	stdlog.SetPrefix("[gogen] ")
	stdlog.SetFlags(0)

	outdir := flag.String("o", "", "Output directory (relative). Required.")
	flag.BoolVar(&quiet, "q", false, "Suppress logging output.")
	flag.BoolVar(&verbose, "v", false, "Increase logging output.")
	noGi := flag.Bool("no-gi", false, "Do not generate a .gitignore in the output directory.")
	help := flag.Bool("h", false, "Print  help.")

	flag.Parse()

	if *help || flag.NArg() < 2 {
		me := filepath.Base(os.Args[0])

		fmt.Fprintf(os.Stderr, "%s usage:\n\t%s [-q] [-h] [--no-gi] -o OUT_DIR IN_PKG [InType=OutType ...]\n\nFor more usage details see https://github.com/mammothbane/gogen/blob/master/README.md\n\nFlags:\n", me, me)
		flag.PrintDefaults()
		os.Exit(2)
	}

	defer func() {
		if r := recover(); r != nil {
			vlog("Caught panic. Exiting to system with error.")
			os.Exit(1)
		}
	}()

	absPkgPath, err := filepath.Abs(*outdir)
	handle(err)
	newPkgName := filepath.Base(absPkgPath)

	typeNames := make(map[string]string, flag.NArg()-1)
	for _, v := range flag.Args()[1:] {
		split := strings.SplitN(v, "=", 2)
		typeNames[split[0]] = split[1]
	}

	var conf loader.Config

	log("Loading and type checking generic code...")
	conf.Import(flag.Arg(0))
	conf.Import(genericsPackageName)
	program, err := conf.Load()
	handle(err)

	log("Determining generic type...")
	genericPkgInfo := program.Package(genericsPackageName)

	gWlkr := genericWalker{
		pkgInfo:     genericPkgInfo,
		genericType: nil,
	}
	for _, fileAst := range genericPkgInfo.Files {
		ast.Walk(&gWlkr, fileAst)

		if gWlkr.genericType != nil {
			break
		}
	}
	if gWlkr.genericType == nil {
		err = fmt.Errorf("Couldn't determine generic type!")
	}
	handle(err)

	pkgInfo := program.Package(flag.Arg(0))

	tdir := filepath.Join(absPkgPath, fmt.Sprintf("gogen-%v", time.Now().Nanosecond()))
	os.MkdirAll(tdir, 0744)
	defer func() {
		vlog("Cleaning up...")
		err := os.RemoveAll(tdir)
		handle(err)
		vlog("Complete.")
	}()

	log("Generating type-specific code...")

	wlkr := walker{typeNames, pkgInfo, gWlkr.genericType}
	var buf bytes.Buffer
	for _, fileAst := range pkgInfo.Files {
		buf.Reset()
		ast.Walk((*typeWalker)(&wlkr), fileAst)
		ast.Walk((*nameWalker)(&wlkr), fileAst)

		fileAst.Name.Name = newPkgName
		err := format.Node(&buf, program.Fset, fileAst)
		handle(err)

		b, err := format.Source(buf.Bytes())
		handle(err)

		fname := filepath.Join(tdir, filepath.Base(program.Fset.Position(fileAst.Pos()).Filename))
		err = ioutil.WriteFile(fname, b, 0744)
		handle(err)
	}

	log("Type-checking generated code...")
	var genConf loader.Config
	genConf.CreateFromFiles(newPkgName, pkgInfo.Files...)
	program, err = genConf.Load()
	handle(err)

	if !*noGi {
		vlog("Creating .gitignore...")
		ioutil.WriteFile(filepath.Join(tdir, ".gitignore"), []byte("*\n"), 0744)
	} else {
		vlog("Skipping .gitignore creation...")
	}

	vlog("Copying generated code to output...")
	f, err := os.Open(tdir)
	handle(err)
	defer f.Close()

	fi, err := f.Readdir(0)
	handle(err)

	for _, v := range fi {
		err := os.Rename(filepath.Join(tdir, v.Name()), filepath.Join(absPkgPath, v.Name()))
		handle(err)
	}
}

func vlog(s string, args ...interface{}) {
	if verbose {
		log(s, args...)
	}
}

func log(s string, args ...interface{}) {
	if !quiet {
		stdlog.Printf(s+"\n", args...)
	}
}

func handle(err error) {
	if err != nil {
		stdlog.Panicf("ERROR: %v", err)
	}
}
