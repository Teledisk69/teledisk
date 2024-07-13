package Handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

func IsPremium(baseUrl string) (bool, error) {

	type User struct {
		is_premium bool
	}
	isPremiumUrl := baseUrl + "/getMe"

	client := &http.Client{}
	res, err := client.Get(isPremiumUrl)
	if err != nil {
		fmt.Println(err)
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	var unJson []User
	json.Unmarshal(body, &unJson)

	for _, c := range unJson {
		if c.is_premium == true {
			return true, nil
		} else {
			return false, nil
		}
	}
	return false, errors.New("error getting user info")
}

func SendDocumentRequest(baseUrl, chat_id, caption, document string) ([]byte, error) {
	sendDocumentUrl := baseUrl + "/sendDocument"

	file, err := os.Open(document)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	defer file.Close()

	fmt.Println(file.Stat())

	pr, pw := io.Pipe()
	defer pr.Close()

	mw := multipart.NewWriter(pw)
	go func() {
		defer pw.Close()
		defer mw.Close()

		p, err := mw.CreateFormFile("document", filepath.Base(file.Name()))
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		_, err = io.Copy(p, file)
		if err != nil {
			pw.CloseWithError(err)
			return
		}
		if err := mw.WriteField("caption", caption); err != nil {
			fmt.Println(err)
		}
		if err := mw.WriteField("chat_id", chat_id); err != nil {
			fmt.Println(err)
		}
	}()

	req, err := http.NewRequest("POST", sendDocumentUrl, pr)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", mw.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("heres body for send document : %v\n", string(body))
	return body, nil
}

type File struct {
	FilePath string `json:"file_path"`
	FileId   string `json:"file_id"`
}

type FileResult struct {
	ResponseOk bool `json:"ok"`
	Result     File `json:"result"`
}

func GetFile(baseUrl, fileId string) FileResult {

	arg := baseUrl + "/getFile"

	data := &bytes.Buffer{}
	writer := multipart.NewWriter(data)
	writerErr := writer.WriteField("file_id", fileId)
	if writerErr != nil {
		log.Fatal(writerErr)
	}

	writer.Close()
	req, err := http.NewRequest("GET", arg, data)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	var bodyResponse FileResult
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	jasonErr := json.Unmarshal(body, &bodyResponse)
	if jasonErr != nil {
		fmt.Println(jasonErr)
	}

	return bodyResponse
}

func ZipNSplit(fileName string) ([]fs.DirEntry, error) {
	fileDirName := "temp"

	var tarArgs = "-cvzf"
	var tarFileName = fileName + ".tar.gz"

	tarCmd := exec.Command("tar", tarArgs, fileDirName+"/"+tarFileName, fileDirName+"/"+fileName)
	if err := tarCmd.Run(); err != nil {
		fmt.Printf("error zipping the file %v", err)
		return nil, err
	}

	var splitArgs = "-b 2000M"
	splitCmd := exec.Command("split", splitArgs, fileDirName+"/"+tarFileName, fileDirName+"/"+tarFileName+".part")
	if err := splitCmd.Run(); err != nil {
		fmt.Printf("error splitting the file %v", err)
		return nil, err
	}
	if err := os.Remove(fileDirName + "/" + fileName); err != nil {
		fmt.Printf("Error deleting file %v", err)
	}
	if err := os.Remove(fileDirName + "/" + tarFileName); err != nil {
		fmt.Printf("Error deleting file %v", err)
	}

	readDir, err := os.ReadDir(fileDirName)
	if err != nil {
		fmt.Printf("Error reading directory %v", err)
	}

	return readDir, nil
}
