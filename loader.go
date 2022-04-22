package memgoloader

import (
	"bytes"
	"cmd/objfile/goobj"
	"cmd/objfile/sys"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"
    "runtime/debug"
    "reflect"
    "log"
    "errors"
)

//go:linkname block runtime.block
func block()

// See reflect/value.go emptyInterface
type interfaceHeader struct {
	typ  unsafe.Pointer
	word unsafe.Pointer
}

// Loader is exported to be used externally
type Loader struct {
	Arch        *sys.Arch
	Goobj       *goobj.Sym
	SymbolPkg   *SymbolPkg
	CodeModules map[string]*CodeModule
    SymPtr      map[string]uintptr
}

// Init initializes the loader structure
func Init(arch *sys.Arch) Loader {
	l := Loader{
		Arch:        arch,
		SymbolPkg:   SymbolPkgInit(),
		CodeModules: make(map[string]*CodeModule),
        SymPtr: make(map[string]uintptr),
	}
	return l
}

// Load loads the bytes from a compiled object
func (l *Loader) Load(plugin *[]byte, pkgName string) error {
	r := bytes.NewReader(*plugin)
	linker, err := ReadObjBytes(r, &pkgName)
	if err != nil {
		return err
	}

	RegAllTypes(l.SymPtr, l.SymbolPkg)
	codeModule, err := Load(linker, l.SymPtr)
	if err != nil {
		return err
	}
	l.CodeModules[pkgName] = codeModule
	return nil
}

func (l *Loader) RegTypes(interfaces ...interface{}){
    RegTypes(l.SymPtr, interfaces)
}

// RegAllTypes registers all necessary symbols
func RegAllTypes(symPtr map[string]uintptr, sp *SymbolPkg) {

	RegTypes(symPtr, time.Duration(0), time.Unix(0, 0))
	RegTypes(symPtr, runtime.LockOSThread)
	// most of time you don't need to register function, but if loader complain about it, you have to.
	RegTypes(symPtr, http.ListenAndServe, http.Dir("/"),
		http.Handler(http.FileServer(http.Dir("/"))), http.FileServer, http.HandleFunc,
		&http.Request{}, &http.Server{})
	RegTypes(symPtr, http.Post, ioutil.ReadAll, &http.Request{}, &http.Response{}, &http.Client{}, bytes.NewBuffer)
	w := sync.WaitGroup{}
	rw := sync.RWMutex{}
	RegTypes(symPtr, &w, w.Wait, &rw)
	symPtr["os.Stdout"] = uintptr(unsafe.Pointer(&os.Stdout))
	RegTypes(symPtr, ioutil.WriteFile)
	RegTypes(symPtr, binary.Read)
	RegTypes(symPtr, bytes.NewReader)
	RegTypes(symPtr, bytes.NewBuffer)
	RegTypes(symPtr, fmt.Println)
	RegTypes(symPtr, exec.Command)

	if sp == nil || len(sp.Syms) == 0 {
		err := RegSymbol(symPtr)
		if err != nil {
			return
		}
		return
	}

	RegSymbolNoOpen(symPtr)

	for _, s := range sp.Syms {
		if strings.HasPrefix(s.Name, "go.itab") {
			//RegItab(symPtr, s.Name, uintptr(s.Address))
		} else {
			if addr, ok := symPtr[s.Name]; ok {
				if addr == 0 {
					symPtr[s.Name] = uintptr(s.Address)
				}
			} else {
				symPtr[s.Name] = uintptr(s.Address)
			}
		}
	}
}

// RegSymbol register common types for relocation
func regBasicSymbol(symPtr map[string]uintptr) {
	int0 := int(0)
	int8d := int8(0)
	int16d := int16(0)
	int32d := int32(0)
	int64d := int64(0)
	RegTypes(symPtr, &int0, &int8d, &int16d, &int32d, &int64d)

	uint0 := uint(0)
	uint8d := uint8(0)
	uint16d := uint16(0)
	uint32d := uint32(0)
	uint64d := uint64(0)
	RegTypes(symPtr, &uint0, &uint8d, &uint16d, &uint32d, &uint64d)

	float32d := float32(0)
	float64d := float64(0)
	complex64d := complex64(0)
	complex128d := complex128(0)
	RegTypes(symPtr, &float32d, &float64d, &complex64d, &complex128d)

	boolTrue := true
	stringEmpty := ""
	unsafePointerd := unsafe.Pointer(&int0)
	uintptrd := uintptr(0)
	RegTypes(symPtr, &boolTrue, &stringEmpty, unsafePointerd, uintptrd)

	RegTypes(symPtr, []int{}, []int8{}, []int16{}, []int32{}, []int64{})
	RegTypes(symPtr, []uint{}, []uint8{}, []uint16{}, []uint32{}, []uint64{})
	RegTypes(symPtr, []float32{}, []float64{}, []complex64{}, []complex128{})
	RegTypes(symPtr, []bool{}, []string{}, []unsafe.Pointer{}, []uintptr{})
    RegFunc(symPtr, "runtime.block", block)
}

// RegSymbolNoOpen registers basic symbols
func RegSymbolNoOpen(symPtr map[string]uintptr) error {
    regBasicSymbol(symPtr)
    return nil
}

// RegItab registers go.itab symbols
func RegItab(symPtr map[string]uintptr, name string, addr uintptr) (err error) {
	symPtr[name] = uintptr(addr)
	bs := strings.TrimLeft(name, "go.itab.")
	bss := strings.Split(bs, ",")
    defer debug.SetPanicOnFault(debug.SetPanicOnFault(true))
    defer func() {
        // recover from panic if one occured. Set err to nil otherwise.
        if (recover() != nil) {
            err = errors.New("Access Violation Ptr Dereference")
            log.Printf("RegItab %s:%x", name, addr)
        }
    }()
	//var slice = sliceHeader{addr, len(bss), len(bss)}
	//ptrs := *(*[]unsafe.Pointer)(unsafe.Pointer(&slice))
    ptrs := *(*[2]uintptr)(unsafe.Pointer(addr))
	for i, ptr := range ptrs {
		typeName := bss[len(bss)-i-1]
		if typeName[0] == '*' {
			var obj interface{} = reflect.TypeOf(0)
			(*interfaceHeader)(unsafe.Pointer(&obj)).word = unsafe.Pointer(ptr)
            if ptr == 0 {
                return errors.New("Invalid pointers")
            }
			typ := obj.(reflect.Type).Elem()
			obj = typ
			typePtr := uintptr((*interfaceHeader)(unsafe.Pointer(&obj)).word)
			symPtr["type."+typeName[1:]] = typePtr
		}
		symPtr["type."+typeName] = uintptr(ptr)
	}
    return nil
}

// RegFunc registers symbols for functions
func RegFunc(symPtr map[string]uintptr, name string, f interface{}) {
	var ptr = GetFuncPtr(f)
	symPtr[name] = ptr
}

// GetFuncPtr gets the pointer to a function interface
func GetFuncPtr(f interface{}) uintptr {
	return *(*uintptr)((*interfaceHeader)(unsafe.Pointer(&f)).word)
}
