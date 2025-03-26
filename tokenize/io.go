package tokenize

import (
	"bytes"
	"encoding/binary"
	"io"
	"log"
	"os"
	"unsafe"
)

func DumpPairs(pairs []Pair, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		log.Fatal("Could not open file")
	}
	defer file.Close()

	err = binary.Write(file, binary.LittleEndian, pairs)
	if err != nil {
		log.Fatal("Write failed")
	}
}

func LoadPairs(filePath string) []Pair {
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Could not open file")
	}
	defer file.Close()

	bin, err := io.ReadAll(file)
	if err != nil {
		log.Fatal("Read failed")
	}
	binReader := bytes.NewReader(bin)

	pairCount := len(bin) / int(unsafe.Sizeof(Pair{}))
	pairs := make([]Pair, pairCount)

	err = binary.Read(binReader, binary.LittleEndian, &pairs)
	if err != nil {
		log.Fatal("Failed to read bytes: ", err)
	}

	return pairs
}
