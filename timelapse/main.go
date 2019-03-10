package main

import (
	"bytes"
	"fmt"
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
	http.HandleFunc("/timelapse",timelapse)
	http.ListenAndServe(":8084",nil)
}

func timelapse(writer http.ResponseWriter, request *http.Request) {
	file, fh, error := request.FormFile("file")
	framerate := request.FormValue("framerate")
	if error != nil {
		http.Error(writer,"Error while uploading",http.StatusBadRequest)
		return
	}
	imagesPath := "/tmp/images"
	if !exists(imagesPath) {
		error := os.MkdirAll(imagesPath, os.ModePerm)
		fmt.Println(error)
	}
	fmt.Println("save file")
	savedFileWithoutExt,savedFileWithExt,_ := saveFile(fh, file, imagesPath)
	fmt.Println("unzip")
	unzipped := unzip(savedFileWithExt,savedFileWithoutExt)
	fmt.Println("create timelapse")
	timelapse := createTimelapse(unzipped,framerate)
	defer file.Close()
	fmt.Println("return")

	bytes, error :=ioutil.ReadFile(timelapse)
	log.Println(error)
	writer.Header().Set("Content-Type", "video/mp4")
	if _, err := writer.Write(bytes); err != nil {
		log.Println("unable to write image.")
	}

}

func unzip(savedFileWithExt string,savedFileWithoutExt string) (string) {
	if !exists(savedFileWithoutExt) {
		os.MkdirAll(savedFileWithoutExt, os.ModePerm)
	}
	unzip := exec.Command("unzip",savedFileWithExt, "-d",savedFileWithoutExt)
	unzip.Start()
	unzip.Wait()
	return savedFileWithoutExt
}

func createTimelapse(savedFileWithoutExt string, framerate string) (string) {
	convert := exec.Command("ffmpeg", "-r", framerate, "-pattern_type", "glob", "-i", "*.png", "-vcodec", "libx264", "timelapse.mp4")
	convert.Dir = savedFileWithoutExt
	convert.Start()
	convert.Wait()
	return savedFileWithoutExt + "/timelapse.mp4"
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