// +build go1.14
// +build !go1.16

package memgoloader

import (
	"cmd/objfile/objabi"
	"strings"
)

// PCDATA and FUNCDATA table indexes.
//
// See funcdata.h and ../cmd/internal/objabi/funcdata.go.
const (
	_PCDATA_RegMapIndex   = 0
	_PCDATA_StackMapIndex = 1
	_PCDATA_InlTreeIndex  = 2

	_FUNCDATA_ArgsPointerMaps    = 0
	_FUNCDATA_LocalsPointerMaps  = 1
	_FUNCDATA_RegPointerMaps     = 2
	_FUNCDATA_StackObjects       = 3
	_FUNCDATA_InlTree            = 4
	_FUNCDATA_OpenCodedDeferInfo = 5
	_ArgsSizeUnknown             = -0x80000000
)

type moduledata struct {
	pclntable    []byte
	ftab         []functab
	filetab      []uint32
	findfunctab  uintptr
	minpc, maxpc uintptr

	text, etext           uintptr
	noptrdata, enoptrdata uintptr
	data, edata           uintptr
	bss, ebss             uintptr
	noptrbss, enoptrbss   uintptr
	end, gcdata, gcbss    uintptr
	types, etypes         uintptr

	textsectmap []textsect
	typelinks   []int32 // offsets from types
	itablinks   []*itab

	ptab []ptabEntry

	pluginpath string
	pkghashes  []modulehash

	modulename   string
	modulehashes []modulehash

	hasmain uint8 // 1 if module contains the main function, 0 otherwise

	gcdatamask, gcbssmask bitvector

	typemap map[typeOff]uintptr // offset to *_rtype in previous module

	bad bool // module failed to load and should be ignored

	next *moduledata
}

// A funcID identifies particular functions that need to be treated
// specially by the runtime.
// Note that in some situations involving plugins, there may be multiple
// copies of a particular special runtime function.
// Note: this list must match the list in cmd/internal/objabi/funcid.go.
type funcID uint8

// Layout of in-memory per-function information prepared by linker
// See https://golang.org/s/go12symtab.
// Keep in sync with linker (../cmd/link/internal/ld/pcln.go:/pclntab)
// and with package debug/gosym and with symtab.go in package runtime.
type _func struct {
	entry   uintptr // start pc
	nameoff int32   // function name

	args        int32  // in/out args size
	deferreturn uint32 // offset of start of a deferreturn call instruction from entry, if any.

	pcsp      int32
	pcfile    int32
	pcln      int32
	npcdata   int32
	funcID    funcID  // set for certain special runtime functions
	_         [2]int8 // unused
	nfuncdata uint8   // must be last
}

func init_func(symbol *ObjSymbol, nameOff, spOff, pcfileOff, pclnOff int) _func {
	fdata := _func{
		entry:       uintptr(0),
		nameoff:     int32(nameOff),
		args:        int32(symbol.Func.Args),
		deferreturn: uint32(0),
		pcsp:        int32(spOff),
		pcfile:      int32(pcfileOff),
		pcln:        int32(pclnOff),
		npcdata:     int32(len(symbol.Func.PCData)),
		funcID:      funcID(objabi.GetFuncID(symbol.Name, strings.TrimLeft(symbol.Func.File[0], FileSymPrefix))),
		nfuncdata:   uint8(len(symbol.Func.FuncData)),
	}
	return fdata
}
