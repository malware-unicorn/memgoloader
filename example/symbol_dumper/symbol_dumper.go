// Copyright 2013 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	//"sort"
	"bytes"
	"cmd/objfile/objfile"
	"compress/zlib"
	"io/ioutil"
)

const helpText = `usage: go tool nm [options] file...
  -n
      an alias for -sort address (numeric),
      for compatibility with other nm commands
  -size
      print symbol size in decimal between address and type
  -sort {address,name,none,size}
      sort output in the given order (default name)
      size orders from largest to smallest
  -type
      print symbol type after name
`

func usage() {
	fmt.Fprintf(os.Stderr, helpText)
	os.Exit(2)
}

var (
	sortOrder = flag.String("sort", "name", "")
	printSize = flag.Bool("size", false, "")
	printType = flag.Bool("type", false, "")

	filePrefix = false
)

func init() {
	flag.Var(nflag(0), "n", "") // alias for -sort address
}

type nflag int

func (nflag) IsBoolFlag() bool {
	return true
}

func (nflag) Set(value string) error {
	if value == "true" {
		*sortOrder = "address"
	}
	return nil
}

func (nflag) String() string {
	if *sortOrder == "address" {
		return "true"
	}
	return "false"
}

func main() {
	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	switch *sortOrder {
	case "address", "name", "none", "size":
		// ok
	default:
		fmt.Fprintf(os.Stderr, "nm: unknown sort order %q\n", *sortOrder)
		os.Exit(2)
	}

	args := flag.Args()
	filePrefix = len(args) > 1
	if len(args) == 0 {
		flag.Usage()
	}

	for _, file := range args {
		nm(file)
	}

	os.Exit(exitCode)
}

var exitCode = 0

func errorf(format string, args ...interface{}) {
	log.Printf(format, args...)
	exitCode = 1
}

func nm(file string) {
	f, err := objfile.Open(file)
	if err != nil {
		errorf("%v", err)
		return
	}
	defer f.Close()

	w := bufio.NewWriter(os.Stdout)

	entries := f.Entries()

	var found bool

	for _, e := range entries {
		syms, err := e.Symbols()
		if err != nil {
			errorf("reading %s: %v", file, err)
		}
		if len(syms) == 0 {
			continue
		}

		found = true

		symbol_file := "./symbols"
		rb := new(bytes.Buffer)
		var count uint32 = 0
		for _, sym := range syms {
			if sym.Code != 'U' && sym.Addr != 0 && len(sym.Name) > 0 {
				count++
			}
		}

		bsize := make([]byte, 4)
		bsize[0] = byte(count)
		bsize[1] = byte((count >> 8) | 0)
		bsize[2] = byte((count >> 16) | 0)
		bsize[3] = byte((count >> 24) | 0)
		rb.Write(bsize)

		for _, sym := range syms {
			if sym.Code != 'U' && sym.Addr != 0 && len(sym.Name) > 0 {
				//fmt.Printf("%8x %s %c\n", sym.Addr, sym.Name, sym.Code)
				b := make([]byte, 4)
				b[0] = byte(sym.Addr)
				b[1] = byte(sym.Addr >> 8)
				b[2] = byte(sym.Addr >> 16)
				b[3] = byte(0)
				rb.Write(b)
			}
		}
		for _, sym := range syms {
			if sym.Code != 'U' && sym.Addr != 0 && len(sym.Name) > 0 {
				//fmt.Printf("%8x %s\n", sym.Addr, sym.Name)
				rb.WriteString(sym.Name)
				rb.WriteString("\x00")
			}
		}
		//ioutil.WriteFile(symbol_file, rb.Bytes(), 0644)
		// Compress
		var zipped bytes.Buffer
		w := zlib.NewWriter(&zipped)
		w.Write(rb.Bytes())
		w.Close()
		ioutil.WriteFile(symbol_file, zipped.Bytes(), 0644)
	}

	if !found {
		errorf("reading %s: no symbols", file)
	}

	w.Flush()
}
