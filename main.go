package main

import (
	"context"

	"github.com/bwz1984/mergefile/merge_file"
)

func main() {
	ctx := context.TODO()
	dstFileName := ""
	srcFileNames := []string{}

	// @todo 错误处理
	merge_file.MergeFile(ctx, srcFileNames, dstFileName)
}
