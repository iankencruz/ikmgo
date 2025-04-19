package utils

import (
	ikmgo "ikm"
	"io/fs"
	"log"
)

func PrintEmbeddedFiles() {
	fs.WalkDir(ikmgo.EmbeddedFiles, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Println("⚠️ Error walking embedded FS:", err)
			return nil
		}
		log.Println("📦 Embedded file:", path)
		return nil
	})
}
