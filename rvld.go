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

	inputFile := linker.NewInputFile(file)
	utils.Assert(len(inputFile.ElfSections) == 11)
}
