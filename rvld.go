package main

import (
	"learn/rvld/pkg/linker"
	"learn/rvld/pkg/utils"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		utils.Fatal("Wrong args")
	}

	file := linker.MustNewFile(os.Args[1])

	objectFile := linker.NewObjectFile(file)
	objectFile.Parse()
	utils.Assert(len(objectFile.ElfSections) == 11)
	utils.Assert(objectFile.FirstGlobal == 10)
	utils.Assert(len(objectFile.ElfSymbols) == 12)

	for _, sym := range objectFile.ElfSymbols {
		println(linker.ElfGetName(objectFile.SymbolStringTable, sym.Name))
	}
}
