package main

import (
	"bytes"
	"fmt"
	"github.com/ddliu/go-httpclient"
	"github.com/satori/go.uuid"
	"io"
	"io/ioutil"
	"log"
	"math"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main()  {
	http.HandleFunc("/resize", resize)
	http.HandleFunc("/resize/percent", resizePercent)
	http.ListenAndServe(":8083",nil)
}

func resizePercent(writer http.ResponseWriter, request *http.Request) {
	file, fh, error := request.FormFile("file")
	percent,_ := strconv.ParseFloat(request.FormValue("percent"),64)
	if error != nil {
		http.Error(writer,"Error while uploading",http.StatusBadRequest)
		return
	}
	imagesPath := "/tmp/images"
	if !exists(imagesPath) {
		os.MkdirAll(imagesPath,os.ModePerm)
	}


	savedFileWithoutExt,savedFileWithExt,fileExtension := saveFile(fh, file, imagesPath)

	res,_ := httpclient.Post("http://exifdata:8082/exifdata/filtered", map[string]string {
		"@file": savedFileWithExt,
		"filter": "Image Height",
	})

	responseAsString, _ := res.ToString()
	temp := strings.TrimSpace(strings.Split(responseAsString,":")[1])
	height, _ := strconv.ParseFloat(temp,64)

	heightPercent := fmt.Sprintf("%f",math.Floor(height * (percent / 100)))
	resized := execResize(savedFileWithoutExt, savedFileWithExt,heightPercent+"x"+heightPercent, fileExtension)
	defer file.Close()


	bytes, error :=ioutil.ReadFile(resized)

	writer.Header().Set("Content-Type", "image/jpeg")
	if _, err := writer.Write(bytes); err != nil {
		log.Println("unable to write image.")
	}

}

func resize(writer http.ResponseWriter, request *http.Request) {
	println("resize request")
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

	savedFileWithoutExt,savedFileWithExt,fileExtension := saveFile(fh, file, imagesPath)
	resized := execResize(savedFileWithoutExt, savedFileWithExt,size, fileExtension)
	defer file.Close()

	bytes, error :=ioutil.ReadFile(resized)
	log.Println(error)
	writer.Header().Set("Content-Type", "image/jpeg")
	if _, err := writer.Write(bytes); err != nil {
		log.Println("unable to write image.")
	}

}

func execResize(savedFileWithoutExt string,savedFileWithExt string, size string,fileExtension string) (string){
	log.Println("size",size)
	convert := exec.Command("convert",savedFileWithExt, "-resize",size,savedFileWithoutExt+"_resized"+fileExtension)
	convert.Start()
	convert.Wait()
	return savedFileWithoutExt+"_resized"+fileExtension
}
func saveFile(fh *multipart.FileHeader, file multipart.File, imagesPath string) (string,string,string) {
	fileExtension := filepath.Ext(fh.Filename)
	uuid, _ := uuid.NewV4()
	savedFileWithExt := imagesPath+"/"+uuid.String()+fileExtension
	savedFileWithoutExt := imagesPath+"/"+uuid.String()
	buf := bytes.NewBuffer(nil)
	io.Copy(buf, file)
	ioutil.WriteFile(savedFileWithExt, buf.Bytes(), os.ModePerm)
	return savedFileWithoutExt,savedFileWithExt,fileExtension
}

// exists returns whether the given file or directory exists
func exists(path string) (bool) {
	_, err := os.Stat(path)
	if err == nil { return true }
	if os.IsNotExist(err)  {return false}
	return true
}