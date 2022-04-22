package memgoloader
import (
    "cmd/objfile/sys"
	"fmt"
	"strings"
    "bytes"
    //"runtime"
)


type Pkg2 struct {
	Syms    map[string]*ObjSymbol
	Arch    string
	PkgPath string
	b       *bytes.Reader
}

func readObjBytes(pkg *Pkg2, linker *Linker) error {
	if pkg.PkgPath == EmptyString {
		pkg.PkgPath = DefaultPkgPath
	}
	if err := pkg.symbols(); err != nil {
		return fmt.Errorf("read error: %v", err)
	}
	if len(linker.Arch) != 0 && linker.Arch != pkg.Arch {
		return fmt.Errorf("read obj error: Arch %s != Arch %s", linker.Arch, pkg.Arch)
	}
    linker.Arch = pkg.Arch

	switch linker.Arch {
	case sys.ArchARM.Name, sys.ArchARM64.Name:
		copy(linker.pclntable, armmoduleHead)
	}
	for _, sym := range pkg.Syms {
		for index, loc := range sym.Reloc {
			sym.Reloc[index].Sym.Name = strings.Replace(loc.Sym.Name, EmptyPkgPath, pkg.PkgPath, -1)
		}
		if sym.Func != nil {
			for index, FuncData := range sym.Func.FuncData {
				sym.Func.FuncData[index] = strings.Replace(FuncData, EmptyPkgPath, pkg.PkgPath, -1)
			}
		}
	}
	for _, sym := range pkg.Syms {
		linker.objsymbolMap[sym.Name] = sym
	}
	return nil
}

// ReadObjBytes exported for external use
func ReadObjBytes(r *bytes.Reader, pkgpath *string) (*Linker, error) {
    linker := initLinker()
	pkg := Pkg2{Syms: make(map[string]*ObjSymbol, 0), b: r, PkgPath: *pkgpath}
	if err := readObjBytes(&pkg, linker); err != nil {
		return nil, err
	}
	if err := linker.addSymbols(); err != nil {
		return nil, err
	}
	return linker, nil
}
