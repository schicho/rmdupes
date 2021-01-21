package main

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

type fileInfo struct {
	filePath string
	checksum string
}

func findFiles(pathToFolder string, filePaths chan string) {
	fileInfos, err := ioutil.ReadDir(pathToFolder)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range fileInfos {
		if !f.IsDir() {
			filePaths <- pathToFolder + string(os.PathSeparator) + f.Name()
		}
	}
	close(filePaths)
}

func deleter(done chan bool, files chan fileInfo, deletedFilesCount *uint32) {
	seenFiles := make(map[string]struct{})
	for file := range files {
		_, ok := seenFiles[file.checksum]
		if !ok { //entry does not exist yet
			seenFiles[file.checksum] = struct{}{}
		} else {
			err := os.Remove(file.filePath)
			if err != nil {
				log.Fatal(err)
			}
			*deletedFilesCount++
		}
	}
	done <- true
}

func hasher(filePaths chan string, files chan fileInfo, wg *sync.WaitGroup) {
	for filePath := range filePaths {
		var checksum, err = hashFileSHA256(filePath)
		if err != nil {
			log.Fatal(err, " Failed hashing: ", filePath)
		}
		files <- fileInfo{filePath, checksum}
	}
	wg.Done()
}

func createHasherPool(hasherCount int, filePaths chan string, files chan fileInfo) {
	var wg sync.WaitGroup
	for i := 0; i < hasherCount; i++ {
		wg.Add(1)
		go hasher(filePaths, files, &wg)
	}
	wg.Wait()
	close(files)
}

func hashFileSHA256(filePath string) (string, error) {
	var sha256Return string
	file, err := os.Open(filePath)
	if err != nil {
		return sha256Return, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return sha256Return, err
	}
	return base64.StdEncoding.EncodeToString(hash.Sum(nil)), nil
}

func RmDupes(pathToFolder string) {
	var deletedFilesCount uint32
	const hasherCount = 20
	filePaths := make(chan string, 50)
	files := make(chan fileInfo, 50)

	go findFiles(pathToFolder, filePaths)
	done := make(chan bool)
	go deleter(done, files, &deletedFilesCount)
	createHasherPool(hasherCount, filePaths, files)
	<-done
	fmt.Println("Deleted a total of", deletedFilesCount, "files.")
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("No directory given. Please enter directory explicitly.\nOr try: rmdupes --help ")
	} else {
		switch os.Args[1] {
		case "--help", "-h":
			fmt.Println("Removes all duplicate files in a directory, based on their SHA256 checksum.")
			fmt.Println("Usage: rmdupes <path to folder>\nFor the current working directory use '.'")
			break
		default:
			RmDupes(os.Args[1])
			break
		}
	}
}
