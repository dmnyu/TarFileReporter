package main

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"github.com/dmnyu/ByteMaths"
	"github.com/google/go-tika/tika"
	"github.com/spf13/cast"
	"io"
	"os"
)

var mediaTypes = map[string]MediaType{}

type MediaType struct {
	Count int
	Size  float64
}

func main() {
	var err error
	server, err := tika.NewServer("tika-server-1.26.jar", "")
	if err != nil {
		panic(err)
	}
	err = server.Start(context.Background())
	if err != nil {
		panic(err)
	}

	client := tika.NewClient(nil, server.URL())
	if err != nil {
		panic(err)
	}

	tarFileLoc := "/home/menneric/Downloads/asiapac5.tgz"
	tarFile, err := os.Open(tarFileLoc)
	if err != nil {
		panic(err)
	}
	fmt.Println(tarFile.Name())
	archive, err := gzip.NewReader(tarFile)

	if err != nil {
		fmt.Println("There is a problem with os.Open")
	}

	tr := tar.NewReader(archive)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			panic(err)
		}

		if hdr.Typeflag == tar.TypeReg {
			tmpLoc := "/home/menneric/Elysium/tmpfile"
			tmpfile, err := os.Create(tmpLoc)
			if err != nil {
				panic(err)
			}
			_, err = io.Copy(tmpfile, tr)
			if err != nil {
				panic(err)
			}
			tmp, err := os.Open(tmpLoc)
			if err != nil {
				panic(err)
			}
			detect, err := client.Detect(context.Background(), tmp)
			if err != nil {
				panic(err)
			}

			fmt.Print(hdr.Name + ",")

			if contains(detect) == true {
				mt := mediaTypes[detect]
				mt.Size += cast.ToFloat64(hdr.Size)
				mt.Count += 1
				mediaTypes[detect] = mt
			} else {
				mediaTypes[detect] = MediaType{1, cast.ToFloat64(hdr.Size)}
			}
		}
	}
	server.Shutdown(context.Background())
	output, _ := os.Create("report.tsv")
	defer output.Close()
	writer := bufio.NewWriter(output)

	for k, v := range mediaTypes {
		writer.WriteString(fmt.Sprintf("%s\t%d\t%s\n", k, v.Count, ByteMaths.ToHuman(v.Size)))
		writer.Flush()
	}
}

func contains(s string) bool {
	for k, _ := range mediaTypes {
		if k == s {
			return true
		}
	}
	return false
}
