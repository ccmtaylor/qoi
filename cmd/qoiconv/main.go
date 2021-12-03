package main

import (
	"fmt"
	"image/png"
	"os"
	"strings"

	"github.com/ccmtaylor/qoi"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Println("usage: " + os.Args[0] + " infile outfile\ninfile being png or qoi")
		return
	}
	infile := os.Args[1]
	outfile := os.Args[2]

	f, err := os.Open(infile)
	if err != nil {
		fmt.Println("Error opening file: ", err)
		return
	}

	img, err := qoi.Decode(f)
	if err != nil {
		fmt.Println("Error decoding file: ", err)
		return
	}

	if !strings.HasSuffix(outfile, ".png") && !strings.HasSuffix(outfile, ".qoi") {
		fmt.Println("Only png or qoi files are supported.")
	}

	of, err := os.Create(outfile)
	if err != nil {
		fmt.Printf("Error creating file: %v", err)
		return
	}
	if strings.HasSuffix(outfile, ".png") {
		png.Encode(of, img)
	} else {
		panic("encode not done yet")
		// qoi.Encode(of, img)
	}
}
