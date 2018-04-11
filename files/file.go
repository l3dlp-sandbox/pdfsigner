package files

import (
	"io"
	"io/ioutil"
	"log"
	"path"
	"path/filepath"
	"strings"

	"bitbucket.org/digitorus/pdfsigner/license"
	"bitbucket.org/digitorus/pdfsigner/signer"
)

func storeTempFile(file io.Reader) (string, error) {
	// TODO: Should we encrypt temporary files?
	tmpFile, err := ioutil.TempFile("", "pdf")
	if err != nil {
		return "", err
	}

	_, err = io.Copy(tmpFile, file)
	if err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func findFilesByPatterns(patterns []string) (matchedFiles []string, err error) {
	for _, f := range patterns {
		m, err := filepath.Glob(f)
		if err != nil {
			return matchedFiles, err
		}
		matchedFiles = append(matchedFiles, m...)
	}
	return matchedFiles, err
}

func SignFilesByPatterns(filePatterns []string, outputPathFlag string, signData signer.SignData) {
	// get files
	files, err := findFilesByPatterns(filePatterns)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		// generate signed file path
		dir, fileName := path.Split(f)
		fileNameArr := strings.Split(fileName, path.Ext(fileName))
		fileNameArr = fileNameArr[:len(fileNameArr)-1]
		fileNameNoExt := strings.Join(fileNameArr, "")
		signedFilePath := path.Join(dir, fileNameNoExt+"_signed"+path.Ext(fileName))

		// sign file
		if err := signer.SignFile(f, signedFilePath, signData); err != nil {
			log.Fatal(err)
		}

		log.Println("Signed PDF file written to: " + signedFilePath)
	}

	license.LD.SaveLimitState()
}
