package merge_file

import (
	"context"
	"testing"
)

func TestMergeFile(t *testing.T) {
	ctx := context.TODO()
	dstFileName := "./test.txt"
	srcFileNames := []string{"./merge_file_test.go", "./merge_file.go", "../go.mod", "../main.go", "../README.md"}

	MergeFile(ctx, srcFileNames, dstFileName)
}
