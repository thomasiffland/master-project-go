package main

import (
	"bytes"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func main()  {
	http.HandleFunc("/exifdata",exifdata)
	http.HandleFunc("/exifdata/filtered",exifdataFiltered)
	http.ListenAndServe(":8082",nil)
}

func exifdata(writer http.ResponseWriter, request *http.Request) {
	file, fh, error := request.FormFile("file")
	if error != nil {
		http.Error(writer,"Error while uploading",http.StatusBadRequest)
		return
	}
	imagesPath := "/tmp/images"
	if !exists(imagesPath) {
		os.MkdirAll(imagesPath,os.ModePerm)
	}

	savedFileWithoutExt,savedFileWithExt := saveFile(fh, file, imagesPath)
	result := execExifdata(savedFileWithoutExt, savedFileWithExt)
	defer file.Close()
	if _, err := writer.Write(result); err != nil {
		log.Println("unable to write image.")
	}

}

func exifdataFiltered(writer http.ResponseWriter, request *http.Request) {
	file, fh, error := request.FormFile("file")
	filter := request.FormValue("filter")
	if error != nil {
		http.Error(writer,"Error while uploading",http.StatusBadRequest)
		return
	}
	imagesPath := "/tmp/images"
	if !exists(imagesPath) {
		os.MkdirAll(imagesPath,os.ModePerm)
	}

	savedFileWithoutExt,savedFileWithExt := saveFile(fh, file, imagesPath)
	result := execExifdataFiltered(savedFileWithoutExt, savedFileWithExt,filter)
	println(result)
	defer file.Close()
	if _, err := writer.Write(result); err != nil {
		log.Println("unable to send response")
	}

}

func execExifdataFiltered(savedFileWithoutExt string,savedFileWithExt string,filter string) ([]byte){
	exiftool := exec.Command("exiftool",savedFileWithExt)
	grep := exec.Command("grep", filter)
	pipe, _ := exiftool.StdoutPipe()
	defer pipe.Close()
	grep.Stdin = pipe
	exiftool.Start()
	res, _ := grep.Output()
	return res
}

func execExifdata(savedFileWithoutExt string,savedFileWithExt string) ([]byte){
	exiftool := exec.Command("exiftool",savedFileWithExt)
	res, _ := exiftool.Output()
	return res
}
func saveFile(fh *multipart.FileHeader, file multipart.File, imagesPath string) (string,string) {
	fileExtension := filepath.Ext(fh.Filename)
	uuid, _ := uuid.NewV4()
	savedFileWithExt := imagesPath+"/"+uuid.String()+"."+fileExtension
	savedFileWithoutExt := imagesPath+"/"+uuid.String()
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, file)
	ioutil.WriteFile(savedFileWithExt, buf.Bytes(), os.ModePerm)
	return savedFileWithoutExt,savedFileWithExt
}

// exists returns whether the given file or directory exists
func exists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil { return true }
	if os.IsNotExist(err)  {return false}
	return true
}