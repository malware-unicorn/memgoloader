package memgoloader

import (
    "bytes"
    "bufio"
    "log"
    "compress/zlib"
    "io"
    "strings"
)

// SymPtr a symbol and its address
type SymPtr struct{
    Name string
    Address int64
}

// SymbolPkg a symbol package
type SymbolPkg struct {
    Syms               []*SymPtr
    Offset             int64
    rd                 *bytes.Reader
}

// SymbolPkgInit initializes symbol packages
func SymbolPkgInit() *SymbolPkg {
    return &SymbolPkg{}
}

// Decompress decompresses zlib compiled object
func Decompress(compressed []byte) ([]byte, error) {
    zr, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		return nil, err
	}
    decompressed := new(bytes.Buffer)
    _, err = io.Copy(decompressed, zr)
    if err != nil {
		return nil, err
	}
    return decompressed.Bytes(), nil
}

// Parse parses the symbols from bytes
func (sp *SymbolPkg) Parse(input []byte) {
    if input == nil {
        log.Fatal("input is nil")
        return
    }
    sp.rd = bytes.NewReader(input)
    // Get num of symbols
    var offsets []int32

    count := int32(sp.readUint32())
    for i := int32(0); i < count; i++ {
        o := int32(sp.readUint32())
        offsets = append(offsets, o)
    }

    for i := int32(0); i < count; i++ {
        name, err := sp.readString('\x00')
        if err != nil {
            log.Printf("Name error")
            name = ""
        }
        //symtab hacks see cmd/link/internal/ld/symtab.go
        switch {
        case strings.HasPrefix(name, "runtime.gcbits.*"):
            name = "runtime.gcbits."
        case strings.HasPrefix(name, "go.string.*"):
            name = "go.string."
        }
        sp.Syms = append(sp.Syms, &SymPtr{Name: name, Address: int64(offsets[i])})
        //s := sp.Syms.Lookup(name, sp.LocalSymVersion)
        //s.Value = int64(offsets[i])
    }
}

func (sp *SymbolPkg) readByte() (byte, error) {
    sp.rd.Seek(sp.Offset, 0)
    sp.Offset++
	return sp.rd.ReadByte()
}

func (sp *SymbolPkg) readUint32() uint32 {
    b0, err := sp.readByte()
    if err != nil {
        log.Fatalln("error reading input: ", err)
    }
    b1, err := sp.readByte()
    if err != nil {
        log.Fatalln("error reading input: ", err)
    }
    b2, err := sp.readByte()
    if err != nil {
        log.Fatalln("error reading input: ", err)
    }
    b3, err := sp.readByte()
    if err != nil {
        log.Fatalln("error reading input: ", err)
    }
    return uint32(b0) | uint32(b1)<<8 | uint32(b2)<<16 | uint32(b3)<<24
}

func (sp *SymbolPkg) readString(delim byte) (string, error) {
    sp.rd.Seek(sp.Offset, 0)
    buffReader := bufio.NewReader(sp.rd)

    line, err := buffReader.ReadSlice(delim)
    sp.Offset += int64(len(line))
    return string(line[:len(line)-1]), err
}
