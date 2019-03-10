package main

import (
	"bytes"
	"github.com/ddliu/go-httpclient"
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
	http.HandleFunc("/rawtojpg",rawtojpg)
	http.HandleFunc("/rawtojpg/grayscale",rawtojpgGrayscale)
	http.ListenAndServe(":8080",nil)
}

func rawtojpgGrayscale(writer http.ResponseWriter, request *http.Request) {
	println("called rawtojpggrayscale")
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
	jpg := execRawToJpg(savedFileWithoutExt, savedFileWithExt)
	defer file.Close()

	res,_ := httpclient.Post("http://localhost:8081/grayscale", map[string]string {
		"@file": jpg,
	})
	bytes, _ := res.ReadAll()


	writer.Header().Set("Content-Type", "image/jpeg")
	if _, err := writer.Write(bytes); err != nil {
		log.Println("unable to write image.")
	}
}

func rawtojpg(writer http.ResponseWriter, request *http.Request) {
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
	jpg := execRawToJpg(savedFileWithoutExt, savedFileWithExt)
	defer file.Close()

	bytes, _ :=ioutil.ReadFile(jpg)
	writer.Header().Set("Content-Type", "image/jpeg")
	if _, err := writer.Write(bytes); err != nil {
		log.Println("unable to write image.")
	}

}

func execRawToJpg(savedFileWithoutExt string,savedFileWithExt string) (string){
	dcraw := exec.Command("dcraw","-c","-w",savedFileWithExt)
	convert := exec.Command("convert", "-",savedFileWithoutExt+".jpg")
	pipe, _ := dcraw.StdoutPipe()
	defer pipe.Close()
	convert.Stdin = pipe
	dcraw.Start()
	convert.Start()
	convert.Wait()
	return savedFileWithoutExt+".jpg"
}
func saveFile(fh *multipart.FileHeader, file multipart.File, imagesPath string) (string,string) {
	fileExtension := filepath.Ext(fh.Filename)
	uuid, _ := uuid.NewV4()
	savedFileWithExt := imagesPath+"/"+uuid.String()+fileExtension
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