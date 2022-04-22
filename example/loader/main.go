package main

import (
	//"cmd/objfile/goobj"
	"flag"
	"fmt"
	"net/http"
	"os"
	//"runtime"
	"strings"
	"sync"
	//"time"
	"unsafe"
	"bytes"
	"github.com/kr/pretty"
	memgoloader "http://github.com/malwareunicorn/memgoloader"
	goparser "http://github.com/malwareunicorn/memgoloader/goparser"
	"io/ioutil"
	"cmd/objfile/sys"
	"log"
	"path/filepath"
	"path"
    "runtime"
)

type arrayFlags struct {
	File    []string
	PkgPath []string
}

func (i *arrayFlags) String() string {
	return "my string representation"
}

func (i *arrayFlags) Set(value string) error {
	s := strings.Split(value, ":")
	i.File = append(i.File, s[0])
	var path string
	if len(s) > 1 {
		path = s[1]
	}
	i.PkgPath = append(i.PkgPath, path)
	return nil
}

func main() {
	var files arrayFlags
	flag.Var(&files, "o", "load go object file")
	var pkgpath = flag.String("p", "", "package path")
	var parseFile = flag.String("parse", "", "parse go object file")
	var run = flag.String("run", "main.main", "run function")
	var times = flag.Int("times", 1, "run count")

	flag.Parse()

	if *parseFile != "" {
		parse(parseFile, pkgpath)
		return
	}

	if len(files.File) == 0 {
		flag.PrintDefaults()
		return
	}
    loader := memgoloader.Init(sys.ArchAMD64)
    for i ,f := range files.File {
        fi, err := os.Open(f)
		if err != nil {
			return
		}
		defer fi.Close()
        data, err := ioutil.ReadAll(fi)
        if err != nil {
            log.Fatal(err)
        }
        files.PkgPath[i] = filepath.Base(path.Dir(f))


        w := sync.WaitGroup{}
        loader.RegTypes(http.ListenAndServe, http.Dir("/"),
		http.Handler(http.FileServer(http.Dir("/"))), http.FileServer, http.HandleFunc,
		&http.Request{}, &http.Server{})
    	loader.RegTypes(runtime.LockOSThread, &w, w.Wait)
    	loader.RegTypes(fmt.Sprint)
        err = loader.Load(&data, files.PkgPath[i])
        if err != nil {
            log.Fatal(err)
        }
        for i, v := range loader.CodeModules["example"].Syms {
            log.Println("%s %s", i, v)
        }
        if cm, ok := loader.CodeModules["example"]; ok {
            runFuncPtr := cm.Syms[*run]
            if runFuncPtr == 0 {
                log.Println("Load error! not find function")
                return
            }
            funcPtrContainer := (uintptr)(unsafe.Pointer(&runFuncPtr))
            runFunc := *(*func())(unsafe.Pointer(&funcPtrContainer))

            var wg sync.WaitGroup
            for j := 0; j < *times; j++ {
                wg.Add(1)
                go func() {
                    runFunc()
                    wg.Done()
                }()
            }
            wg.Wait()
            cm.Unload()
        } else {
            log.Println("Empty")
        }
        os.Stdout.Sync()
    }

}

func parse(file, pkgpath *string) {
    log.Printf("Parse")
	if *file == "" {
		flag.PrintDefaults()
		return
	}

	f, err := os.Open(*file)
	if err != nil {
		fmt.Printf("%# v\n", err)
		return
	}
    data, err := ioutil.ReadAll(f)
    if err != nil {
        log.Fatal(err)
    }
    bytesReader := bytes.NewReader(data)
    //var loader * memgoloader.Loader =
    memgoloader.Init(sys.ArchAMD64)
	obj, err := goparser.Parse(bytesReader, *pkgpath)
	pretty.Printf("%# v\n", obj)
	f.Close()
	if err != nil {
		fmt.Printf("error reading %s: %v\n", *file, err)
		return
	}
}
