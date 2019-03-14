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
	http.HandleFunc("/grayscale",grayscale)
	http.HandleFunc("/grayscale/resize",grayscaleResize)
	http.ListenAndServe(":8081",nil)
}

func grayscaleResize(writer http.ResponseWriter, request *http.Request) {
	file, fh, error := request.FormFile("file")
	size := request.FormValue("size")
	if error != nil {
		http.Error(writer,"Error while uploading",http.StatusBadRequest)
		return
	}
	imagesPath := "/tmp/images"
	if !exists(imagesPath) {
		os.MkdirAll(imagesPath,os.ModePerm)
	}

	savedFileWithoutExt,savedFileWithExt := saveFile(fh, file, imagesPath)
	grayscale := execGrayscale(savedFileWithoutExt, savedFileWithExt)
	defer file.Close()

	res,error := httpclient.Post("http://resize:8083/resize", map[string]string {
		"@file": grayscale,
		"size": size,
	})
	println(res)
	bytes, _ :=res.ReadAll()
	writer.Header().Set("Content-Type", "image/jpeg")
	if _, err := writer.Write(bytes); err != nil {
		log.Println("unable to write image.")
	}

}

func grayscale(writer http.ResponseWriter, request *http.Request) {
	println("grayscale called")
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
	grayscale := execGrayscale(savedFileWithoutExt, savedFileWithExt)
	defer file.Close()

	bytes, _ :=ioutil.ReadFile(grayscale)

	writer.Header().Set("Content-Type", "image/jpeg")
	if _, err := writer.Write(bytes); err != nil {
		log.Println("unable to write image.")
	}

}

func execGrayscale(savedFileWithoutExt string,savedFileWithExt string) (string){
	convert := exec.Command("convert",savedFileWithExt, "-colorspace","Gray",savedFileWithoutExt+"_grayscale.jpg")
	convert.Start()
	convert.Wait()
	return savedFileWithoutExt+"_grayscale.jpg"
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