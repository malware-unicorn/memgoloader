// +build go1.8
// +build !go1.16

package memgoloader

import (
    goparser "git.redteam.fb/malwareunicorn/memgoloader/goparser"
	"fmt"
)

func (pkg *Pkg2) symbols() error {
    obj, err := goparser.Parse(pkg.b, pkg.PkgPath)
	if err != nil {
		return fmt.Errorf("read error: %v", err)
	}
	pkg.Arch = obj.Arch
	fd := readAtSeeker{ReadSeeker: pkg.b}
	for _, sym := range obj.Syms {
		symbol := &ObjSymbol{}
		symbol.Name = sym.Name
		symbol.Kind = int(sym.Kind)
		symbol.DupOK = sym.DupOK
		symbol.Size = int64(sym.Size)
		symbol.Data, err = fd.BytesAt(sym.Data.Offset, sym.Data.Size)
		if err != nil {
			return fmt.Errorf("read error: %v", err)
		}
		grow(&symbol.Data, (int)(symbol.Size))
		for _, loc := range sym.Reloc {
			reloc := Reloc{
				Offset: int(loc.Offset),
				Sym:    &Sym{Name: loc.Sym.Name, Offset: InvalidOffset},
				Type:   int(loc.Type),
				Size:   int(loc.Size),
				Add:    int(loc.Add)}
			symbol.Reloc = append(symbol.Reloc, reloc)
		}
		if sym.Func != nil {
			symbol.Func = &FuncInfo{}
			symbol.Func.Args = uint32(sym.Func.Args)
			symbol.Func.File = sym.Func.File
			symbol.Func.PCSP, err = fd.BytesAt(sym.Func.PCSP.Offset, sym.Func.PCSP.Size)
			if err != nil {
				return fmt.Errorf("read error: %v", err)
			}
			symbol.Func.PCFile, err = fd.BytesAt(sym.Func.PCFile.Offset, sym.Func.PCFile.Size)
			if err != nil {
				return fmt.Errorf("read error: %v", err)
			}
			symbol.Func.PCLine, err = fd.BytesAt(sym.Func.PCLine.Offset, sym.Func.PCLine.Size)
			if err != nil {
				return fmt.Errorf("read error: %v", err)
			}

			for _, data := range sym.Func.PCData {
				pcdata, err := fd.BytesAt(data.Offset, data.Size)
				if err != nil {
					return fmt.Errorf("read error: %v", err)
				}
				symbol.Func.PCData = append(symbol.Func.PCData, pcdata)
			}
			for _, data := range sym.Func.FuncData {
				symbol.Func.FuncData = append(symbol.Func.FuncData, data.Sym.Name)
			}

			//if err = initInline(sym.Func, symbol.Func, pkg.PkgPath, &fd); err != nil {
			//	return fmt.Errorf("read error: %v", err)
			//}
		}
		pkg.Syms[sym.Name] = symbol
	}
	return nil
}
