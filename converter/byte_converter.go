package main

import (
	"fmt"
	"io"
	"os"
)

const icoFilePath = `C:\Users\Emir\GolandProjects\Luminous\icon\switch.ico`

func main() {

	arrayName := "Icon"
	packageName := "icon"

	file, err := os.Open(icoFilePath)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {

		}
	}(file)

	fmt.Printf("package %s\n\n", packageName)
	fmt.Printf("var %s []byte = []byte{", arrayName)

	buf := make([]byte, 1)
	var totalBytes uint64
	var n int
	for n, err = file.Read(buf); n > 0 && err == nil; {
		if totalBytes%12 == 0 {
			fmt.Printf("\n\t")
		}
		fmt.Printf("0x%02x, ", buf[0])
		totalBytes++
		n, err = file.Read(buf)
	}
	if err != nil && err != io.EOF {
		_ = fmt.Errorf("error reading file: %v", err)
		return
	}
	fmt.Print("\n}\n\n")
}
