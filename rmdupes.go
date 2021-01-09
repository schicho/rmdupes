package main

import (
	"encoding/base64"
	"fmt"
	"crypto/sha256"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"
)

var filePaths = make(chan string, 50)
var files = make(chan fileWithHash, 50)

type fileWithHash struct {
	filePath string
	checksum string
}

func findFiles(pathToFolder string) {
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

func deleter(done chan bool, deletedFilesCount *uint32) {
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

func hasher(wg *sync.WaitGroup) {
	for filePath := range filePaths {
		var checksum, err = hashFileSHA256(filePath)
		if err != nil {
			log.Fatal(err, " Failed hashing: ", filePath)
		}
		files <- fileWithHash{filePath, checksum}
	}
	wg.Done()
}

func createHasherPool(hasherCount int) {
	var wg sync.WaitGroup
	for i := 0; i < hasherCount; i++ {
		wg.Add(1)
		go hasher(&wg)
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
	go findFiles(pathToFolder)
	done := make(chan bool)
	go deleter(done, &deletedFilesCount)
	hasherCount := 10
	createHasherPool(hasherCount)
	<-done
	fmt.Println("Deleted a total of", deletedFilesCount, "files.")
}

func main() {
	if len(os.Args) == 1 {
		fmt.Println("No directory given. Please enter directory explicitly.\nOr try: rmdupes --help ")
	} else {
		switch os.Args[1] {
		case "--help":
		case "-h":
			fmt.Println("Removes all duplicate files in a directory, based on their CRC32 checksum.")
			fmt.Println("Usage: rmdupes <path to folder>\nFor the current working directory use '.'")
			break
		default:
			RmDupes(os.Args[1])
			break
		}
	}
}
