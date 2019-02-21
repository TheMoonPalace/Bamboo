package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	zip2 "github.com/alexmullins/zip"
)

var (
	zipEnableName 	string
	zipSourcePath 	string
	zipPassword		string
	zipOutPath		string
)

func init() {
	flag.StringVar(&zipEnableName, "r", "false", "enable random zip nameing")
	flag.StringVar(&zipSourcePath, "s", "", "files source path")
	flag.StringVar(&zipPassword, "k", "", "zip file password")
	flag.StringVar(&zipOutPath, "o", "out", "out file path")

	flag.Parse()
}

func main() {
	var e error
	var confirm string

	if zipSourcePath == "" {
		flag.Usage(); return
	}

	// 格式化源目录地址
	zipSourcePath, e = filepath.Abs(zipSourcePath)
	if e != nil {
		log.Fatalln(e); return
	}

	// 格式化输出目录地址
	zipOutPath, e = filepath.Abs(zipOutPath)
	if e != nil {
		log.Fatalln(e); return
	}

	// 向用户确认
	fmt.Println("Confirm Source Path: [", zipSourcePath, "]，Password: [", zipPassword ,"]? [Y/N]")
	_, e = fmt.Scanln(&confirm)
	if e != nil {
		log.Fatalln(e); return
	}

	if strings.ToLower(confirm) == "y" {
		bale()
	}
}

var (
	recycle  []string
	dirName  map[string]string
)

// bale. 打包
func bale() {
	reName()
	compress()
	clean()
}

// reName. 复制并重命名
func reName() {
	log.Println("begin copy files")

	dirName = make(map[string]string)

	// 成功copy数量
	number := 0
	// 总数量
	totalNumber := 0

	e := filepath.Walk(zipSourcePath, func (fName string, fInfo os.FileInfo, err error) error {
		if fName == zipSourcePath {
			return nil
		}

		totalNumber++

		if err != nil {
			return err
		}

		err = copyAndRename(fName, fInfo)

		if err != nil {
			return err
		}

		if !fInfo.IsDir() {
			number++
		}

		return nil
	})

	if e != nil {
		log.Fatalln(e)
	}

	log.Println("begin copy complete, total:", totalNumber, ", success:", number)
}

func compress() {
	log.Println("begin compress")

	f, e := ioutil.ReadDir(zipOutPath)
	if e != nil {
		log.Println("ReadDir Fail:", e); return
	}

	suffix := ".zip"

	for _, dir := range f {
		zipName, e := filepath.Abs(fmt.Sprintf("%v/%v%v", zipOutPath, dir.Name(), suffix))
		if e != nil {
			log.Println("= Path Fail:", e); return
		}

		zip, e := os.Create(zipName)
		if e != nil {
			log.Println("= Create Fail:", e); return
		}

		fileRelName, e := filepath.Abs(zipOutPath + "/" + dir.Name())
		if e != nil {
			log.Println("= Open Fail:", e); return
		}

		zipWriter := zip2.NewWriter(zip)

		if dir.IsDir() {
			filepath.Walk(fileRelName, func (fName string, fInfo os.FileInfo, err error) error {
				return compressFile(fName, fInfo, zipWriter)
			})
		} else {
			compressFile(fileRelName, dir, zipWriter)
		}

		zipWriter.Close()
		recycle = append(recycle, fileRelName)
	}

	log.Println("compress complete")
}

// compressFile. 压缩
func compressFile(fName string, fInfo os.FileInfo, zipWriter *zip2.Writer) error {
	if fInfo.IsDir() {
		return nil
	}

	file, e := os.Open(fName)
	if e != nil {
		log.Println("Walk Fail:", e); return e
	}

	zw, e := zipWriter.Encrypt(fInfo.Name(), zipPassword)
	if e != nil {
		log.Println("Encrypt Fail:", e); return e
	}

	_, e = io.Copy(zw, file)
	if e != nil {
		log.Println("Copy Fail:", e); return e
	}

	zipWriter.Flush()
	file.Close()

	return nil
}

// copyAndRename.
func copyAndRename(fName string, fInfo os.FileInfo) error {
	if fInfo.IsDir() {
		if zipEnableName == "true" {
			dirName[fName] = strconv.FormatInt(getUid(), 10)
		} else {
			dirName[fName] = fInfo.Name()
		}

		return nil
	}

	var e error
	var s = filepath.Ext(fName)
	var n = strings.TrimRight(fInfo.Name(), s)

	oldDirName := strings.Replace(fName, fInfo.Name(), "", -1)
	oldDirName  = strings.TrimRight(oldDirName, "\\")
	oldDirName  = strings.TrimRight(oldDirName, "/")

	if _, ok := dirName[oldDirName]; !ok {
		log.Println("Path Error: ", oldDirName, " Not In dirName[]"); return nil
	}

	if zipEnableName == "true" {
		n = strconv.FormatInt(getUid(), 10)
	}

	newDirName, e := filepath.Abs(fmt.Sprintf("%v/%v", zipOutPath, dirName[oldDirName]))
	newFileName   := fmt.Sprintf("%v/%v.%v", dirName[oldDirName], n, s)
	newFileName, e = filepath.Abs(zipOutPath + "/" + newFileName)

	if e != nil {
		log.Println("Path Error:", e); return e
	}

	// 判断目录是否存在
	_, e = os.Stat(newDirName)
	if os.IsNotExist(e) {
		os.MkdirAll(newDirName, 0666)
	}

	src, e := os.Open(fName)
	if e != nil {
		log.Println("Open Fail:", e); return e
	}

	dst, e := os.OpenFile(newFileName, os.O_WRONLY|os.O_CREATE, 0666)
	if e != nil {
		log.Println("Open Fail:", e); return e
	}

	_, e = io.Copy(dst, src)
	if e != nil {
		log.Println("Copy Fail:", e); return e
	}

	src.Close()
	dst.Close()

	return nil
}

// clean. 清扫旧文件
func clean() {
	log.Println("begin clean")

	for _, filePath := range recycle {
	reload:

		e := os.RemoveAll(filePath)
		if e != nil {
			goto reload
		}
	}

	log.Println("clean end")
}

// 使用过的id列表
var uuiDs []int64

// inSliceInt64. 判断数组是否存在该元素
func inSliceInt64(val int64) bool {
	for _, v := range uuiDs {
		if val == v {
			return true
		}
	}

	return false
}

// GetUid. 返回一个唯一的（伪）int64 id
func getUid() int64 {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano() + int64(rand.Intn(100))))

	t := time.Now().UnixNano()
	r := t + int64(rnd.Intn(99999 - 10000) + 10000)

	if inSliceInt64(r) {
		r = getUid()
	} else {
		uuiDs = append(uuiDs, r)
	}

	return r
}

