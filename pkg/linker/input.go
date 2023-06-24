package linker

import "learn/rvld/pkg/utils"

func ReadInputFiles(ctx *Context, remainings []string) {
	for _, arg := range remainings {
		var ok bool
		if arg, ok = utils.RemovePrefix(arg, "-l"); ok {
			ReadFile(ctx, FindLibrary(ctx, arg))
		} else {
			ReadFile(ctx, MustNewFile(arg))
		}
	}
}

func ReadFile(ctx *Context, file *File) {
	ft := GetFileType(file.Contents)
	switch ft {
	case FileTypeObject:
		ctx.Objs = append(ctx.Objs, CreateObjectFile(ctx, file, false))
	case FileTypeArchive:
		// 静态链接库是归档文件
		// 静态链接库文件就是由多个 obj 文件打包到一起的归档文件
		for _, child := range ReadArchiveMembers(file) {
			utils.Assert(GetFileType(child.Contents) == FileTypeObject)
			ctx.Objs = append(ctx.Objs, CreateObjectFile(ctx, child, true))
		}
	default:
		utils.Fatal("Unknown file type")
	}
}

func CreateObjectFile(ctx *Context, file *File, inLib bool) *ObjectFile {
	CheckFileCompatibility(ctx, file)
	obj := NewObjectFile(file, !inLib)
	obj.Parse(ctx)
	return obj
}
