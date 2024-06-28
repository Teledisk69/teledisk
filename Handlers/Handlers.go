package Handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)


func IsPremium(baseUrl string) (bool, error){

  type User struct{
  is_premium bool
}
  isPremiumUrl := baseUrl + "/getMe"

  client := &http.Client{}
  res,err := client.Get(isPremiumUrl)
  if err != nil{
    fmt.Println(err)
  }
  defer res.Body.Close()
  body, _ := io.ReadAll(res.Body)


  var unJson []User
  json.Unmarshal(body, &unJson)

  for _, c := range unJson{
    if c.is_premium == true{
    return true, nil
    }else{
      return false, nil
    }
  }
  return  false,errors.New("error getting user info")
}

func SendDocumentRequest(baseUrl,chat_id, document string) string{
  sendDocumentUrl := baseUrl + "/sendDocument"

  file,err := os.Open(document)
  if err != nil{
    log.Fatal(err)
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


  chatidErr := mw.WriteField("chat_id", chat_id)
  if chatidErr != nil{
    log.Fatal(chatidErr)
  }
    }()


  req,err := http.NewRequest("POST", sendDocumentUrl, pr)
  if err != nil{
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
    return string(body)
}


type File struct{
    FilePath string `json:"file_path"`
    FileId string `json:"file_id"`
  }

type FileResult struct {
    ResponseOk bool `json:"ok"`
    Result File `json:"result"`
}



func GetFile(baseUrl, fileId string) FileResult{

  arg :=  baseUrl + "/getFile"
  
  data := &bytes.Buffer{}
  writer := multipart.NewWriter(data)
  writerErr := writer.WriteField("file_id", fileId)
  if writerErr != nil{
    log.Fatal(writerErr)
  }

  writer.Close()
  req,err := http.NewRequest("GET", arg ,data)
  if err != nil{
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
  body,_ := io.ReadAll(resp.Body)
  jasonErr := json.Unmarshal(body, &bodyResponse);if jasonErr != nil{fmt.Println(jasonErr)}

  return bodyResponse


}

