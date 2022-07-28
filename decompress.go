package tlsclient

import (
	"compress/gzip"
	"io/ioutil"
	"strings"

	"github.com/andybalholm/brotli"
)

func BodyDecompress(body, encoding string) (decompressedData []byte) {

	if encoding == "gzip" {

		data := strings.NewReader(string(body))
		reader, _ := gzip.NewReader(data)
		decompressedData, _ = ioutil.ReadAll(reader)

		return decompressedData

	} else if encoding == "br" {

		data := strings.NewReader(string(body))
		reader := brotli.NewReader(data)
		decompressedData, _ = ioutil.ReadAll(reader)
		return decompressedData

	} else {
		return []byte("")
	}

}
