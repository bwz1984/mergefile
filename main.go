package main

import (
	"bufio"
	"context"
	"log"
	"os"
	"sync"
)

func (mfs *mergeFileStruct) producer(ctx context.Context, srcFileNames []string) {
	wg := mfs.wg
	lineChan := mfs.lineChan
	wg.Add(len(srcFileNames))
	srcFileCnt := len(srcFileNames)
	for _, srcFileName := range srcFileNames {
		go func(fileName string) {
			defer func() {
				wg.Done()
				mfs.fileReadDoneCnt += 1
				if mfs.fileReadDoneCnt == srcFileCnt {
					close(lineChan)
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

func (mfs *mergeFileStruct) consumer(ctx context.Context, dstFileName string) {
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
			if !ok {
				break
			}
			dstFile.WriteString(lineStr + "\n")
		}
	}()
}

type mergeFileStruct struct {
	wg              *sync.WaitGroup // 阻塞主进程 等待所有goroutine退出
	lineChan        chan string     // 传递文件行数据
	fileReadDoneCnt int             // 文件读取完成数量
}

func main() {
	ctx := context.TODO()
	dstFileName := ""
	srcFileNames := []string{}

	wg := &sync.WaitGroup{}
	lineChan := make(chan string, 1024*1024) // 设置缓冲区大小 需要斟酌 保证一直有剩余空间可写

	mfs := &mergeFileStruct{
		wg:              wg,
		lineChan:        lineChan,
		fileReadDoneCnt: 0,
	}

	mfs.consumer(ctx, dstFileName)

	mfs.producer(ctx, srcFileNames)

	wg.Wait()
}
