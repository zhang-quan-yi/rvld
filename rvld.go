package main

import (
	"fmt"
	"learn/rvld/pkg/linker"
	"learn/rvld/pkg/utils"
	"os"
	"strings"
)

var version string

func main() {
	ctx := linker.NewContext()
	remaining := parseArgs(ctx)

	if ctx.Args.Emulation == linker.MachineTypeNone {
		// 没有解析到 -m 参数。
		// 需要从第一个 object 文件中读取 Emulation 的值
		for _, filename := range remaining {
			if strings.HasPrefix(filename, "-") {
				continue
			}

			file := linker.MustNewFile(filename)
			ctx.Args.Emulation = linker.GetMachineTypeFromContents(file.Contents)
			if ctx.Args.Emulation != linker.MachineTypeNone {
				break
			}
		}
	}

	if ctx.Args.Emulation != linker.MachineTypeRISCV64 {
		utils.Fatal("Unknown emulation type")
	}

	linker.ReadInputFiles(ctx, remaining)
	linker.ResolveSymbols(ctx)

	for _, o := range ctx.Objs {
		if o.File.Name == "out/tests/hello/a.o" {
			for _, symbol := range o.Symbols {
				println(symbol.Name)
			}
		}
	}
}

func parseArgs(ctx *linker.Context) []string {
	args := os.Args[1:]

	dashes := func(name string) []string {
		if len(name) == 1 {
			return []string{"-" + name}
		}
		return []string{"-" + name, "--" + name}
	}

	arg := ""
	readArg := func(name string) bool {
		for _, opt := range dashes(name) {
			if args[0] == opt {
				if len(args) == 1 {
					utils.Fatal(fmt.Sprintf("Option -%s: argument missing", name))
				}
				arg = args[1]
				args = args[2:]
				return true
			}

			prefix := opt
			if len(name) > 1 {
				prefix += "="
			}
			if strings.HasPrefix(args[0], prefix) {
				arg = args[0][len(prefix):]
				args = args[1:]
				return true
			}
		}
		return false
	}

	readFlag := func(name string) bool {
		for _, opt := range dashes(name) {
			if args[0] == opt {
				args = args[1:]
				return true
			}
		}
		return false
	}

	remaining := make([]string, 0)
	for len(args) > 0 {
		if readFlag("help") {
			fmt.Printf("Usage: %s [options] file...\n", os.Args[0])
			os.Exit(0)
		}
		if readArg("o") || readArg("output") {
			ctx.Args.Output = arg
		} else if readFlag("v") || readFlag("version") {
			fmt.Printf("rvld %s\n", version)
			os.Exit(0)
		} else if readArg("m") {
			if arg == "elf64lriscv" {
				ctx.Args.Emulation = linker.MachineTypeRISCV64
			} else {
				utils.Fatal(fmt.Sprintf("Unknown -m argument: %s", arg))
			}
		} else if readArg("L") {
			// L 表示 library path
			// 比如 . /usr/lib/gcc...
			ctx.Args.LibraryPaths = append(ctx.Args.LibraryPaths, arg)
		} else if readArg("l") {
			// l 表示静态链接库文件
			// 去 library path 里去找这些库文件
			remaining = append(remaining, "-l"+arg)
		} else if readArg("sysroot") || readFlag("static") || readArg("plugin") || readArg("plugin-opt") || readFlag("as-needed") || readFlag("start-group") || readFlag("end-group") || readArg("hash-style") || readArg("build-id") || readFlag("s") || readFlag("no-relax") {
			// Ignored
		} else {
			if args[0][0] == '-' {
				utils.Fatal(fmt.Sprintf("Unknown command line options: %s", args[0]))
			}
			remaining = append(remaining, args[0])
			args = args[1:]
		}
	}
	return remaining
}
