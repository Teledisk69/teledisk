package Methods

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"tg/telegram-storage/Handlers"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

// will be splitting here if the size is more than 2gb else it will be saved directly
func UploadFile(c echo.Context) error {

	type File struct {
		FileName string `json:"file_name"`
		FileSize int    `json:"file_size"`
		FileId   string `json:"file_id"`
	}

	type FileResponse struct {
		Document File `json:"document"`
	}

	type uploadResponse struct {
		Ok     bool         `json:"ok"`
		Result FileResponse `json:"result"`
	}

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}
	baseUrl := os.Getenv("BASE_URL")

	//use multipart reader here
	newRequest := c.Request()
	newMultiReader, mrErr := newRequest.MultipartReader()
	if mrErr != nil {
		fmt.Println(mrErr)
		return mrErr
	}

	fileDirName := "temp"
	mkdirErr := os.Mkdir(fileDirName, 0750)
	if mkdirErr != nil {
		fmt.Println("dir already exist")
	}

	var chatId string
	var uploader uploadResponse
	for {
		part, pErr := newMultiReader.NextPart()
		if pErr == io.EOF {
			break
		}

		if part.FormName() == "chat_id" {
			buf := new(bytes.Buffer)
			buf.ReadFrom(part)
			chatId = buf.String()
			continue
		}

		if part.FileName() == "file" {
			continue
		}

		dst, err := os.Create(fileDirName + "/" + part.FileName())
		if err != nil {
			fmt.Println("Error creating file:", err)
			return err
		}
		defer dst.Close()

		_, err = io.Copy(dst, part)
		if err != nil {
			fmt.Println("Error copying file:", err)
			return err
		}
		document, _ := dst.Stat()
		var fileSize int64 = document.Size()

		const fileChunk = 2097152000        // 2gb split size
		const fileChunkPremium = 4097152000 // 4gb split for premium user
		isPremium, err := Handlers.IsPremium(baseUrl)
		if err != nil {
			fmt.Println(err)
		}

    openDocument,err := os.Open(fileDirName + "/" + part.FileName())
    if err != nil{
      fmt.Printf("ERROR OPENING FILE %V", err )
    }

		if isPremium && fileSize > fileChunkPremium || fileSize > fileChunk {
			fmt.Printf("the file size is : %v", fileSize)

			totalPartsNum := uint64(math.Ceil(float64(fileSize) / float64(fileChunk)))
			for i := uint64(0); i < totalPartsNum; i++ {
				partSize := int(math.Min(fileChunk, float64(fileSize-int64(i*fileChunk))))
				partBuffer := make([]byte, partSize)
        if _, err := openDocument.Read(partBuffer);err != nil {fmt.Printf("ERROR READING BUFFER %v", err)}
				// write to disk
				fileName := strconv.FormatUint(i, 10) + "_" + part.FileName()
				_, err := os.Create(fileDirName + "/" + fileName)
				if err != nil {
					fmt.Println("error while creating file :", err)
					os.Exit(1)
				}
				// write/save buffer to disk
				if err := os.WriteFile(fileDirName+"/"+fileName, partBuffer, os.ModeAppend); err != nil {
					fmt.Printf("error writing file %v", err)
				}
				fmt.Println("Split to : ", fileName)

				//for _, files := range readDir {
				fmt.Println("uploading file :" + fileName)
				sendFile := Handlers.SendDocumentRequest(baseUrl, chatId, fileDirName+"/"+fileName)
				json.Unmarshal([]byte(sendFile), &uploader)
				fmt.Println(uploader)
				//err := os.RemoveAll(files.Name())
				if err != nil {
					fmt.Println(err)
				}
			}

		} else {
			fmt.Printf("the file size is %d", fileSize)
			sendFile := Handlers.SendDocumentRequest(baseUrl, chatId, fileDirName+"/"+part.FileName())
			json.Unmarshal([]byte(sendFile), &uploader)
			fmt.Println(uploader)
			//err := os.RemoveAll(part.FileName())
			if err != nil {
				fmt.Println(err)
			}
		}
	}
	return c.JSON(http.StatusOK, uploader)
}

