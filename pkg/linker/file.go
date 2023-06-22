package linker

import (
	"learn/rvld/pkg/utils"
	"os"
)

type File struct {
	Name     string
	Contents []byte
	Parent   *File
}

func MustNewFile(filename string) *File {
	contents, err := os.ReadFile(filename)
	utils.MustNo(err)
	return &File{
		Name:     filename,
		Contents: contents,
	}
}

func OpenLibrary(filepath string) *File {
	contents, err := os.ReadFile(filepath)
	if err != nil {
		return nil
	}
	return &File{
		Name:     filepath,
		Contents: contents,
	}
}

func FindLibrary(ctx *Context, name string) *File {
	for _, dir := range ctx.Args.LibraryPaths {
		libraryFileName := dir + "/lib" + name + ".a"
		if file := OpenLibrary(libraryFileName); file != nil {
			return file
		}
	}
	utils.Fatal("Library not found!")
	return nil
}
