package merge_file

import (
	"bufio"
	"context"
	"log"
	"os"
	"sync"
	"sync/atomic"
)

type mergeFileStruct struct {
	wg              *sync.WaitGroup // 阻塞主进程 等待所有goroutine退出
	lineChan        chan string     // 传递文件行数据
	fileReadDoneCnt int32           // 文件读取完成数量
}

func (mfs *mergeFileStruct) produce(ctx context.Context, srcFileNames []string) {
	wg := mfs.wg
	lineChan := mfs.lineChan
	wg.Add(len(srcFileNames))
	srcFileCnt := int32(len(srcFileNames))
	for _, srcFileName := range srcFileNames {
		go func(fileName string) {
			defer func() {
				wg.Done()
				atomic.AddInt32(&mfs.fileReadDoneCnt, 1) // 并发安全
				if mfs.fileReadDoneCnt == srcFileCnt {
					close(lineChan) // 所有文件读取完成 关闭通道 可继续从通道读取数据
				}
			}()
			file, err := os.Open(fileName) // 打开文件
			if err != nil {
				log.Printf("os.Open %v fail, err:%v", fileName, err)
				return
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				lineChan <- scanner.Text() // 按行写入
			}
		}(srcFileName)
	}
}

func (mfs *mergeFileStruct) consume(ctx context.Context, dstFileName string) {
	wg := mfs.wg
	lineChan := mfs.lineChan
	wg.Add(1)
	go func() {
		defer wg.Done()
		dstFile, err := os.Create(dstFileName)
		if err != nil {
			log.Printf("os.Create %v fail, err:%v", dstFileName, err)
			return
		}
		defer dstFile.Close()
		for {
			lineStr, ok := <-lineChan
			if !ok { // 已关闭，且无数据可读
				break
			}
			dstFile.WriteString(lineStr + "\n")
		}
	}()
}

func MergeFile(ctx context.Context, srcFileNames []string, dstFileName string) {
	wg := &sync.WaitGroup{}                  // 等待读写goroutine读写正常退出
	lineChan := make(chan string, 1024*1024) // 设置缓冲区大小 需要斟酌 保证一直有剩余空间可写

	mfs := &mergeFileStruct{
		wg:              wg,
		lineChan:        lineChan,
		fileReadDoneCnt: 0,
	}

	mfs.consume(ctx, dstFileName)

	mfs.produce(ctx, srcFileNames)

	wg.Wait()
}
