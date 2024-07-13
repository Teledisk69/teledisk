package Methods

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"tg/telegram-storage/Handlers"
)

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

	//loading baseurl env

	if err := godotenv.Load(); err != nil {
		log.Panic("Error loading .env file")
		return c.JSON(http.StatusGone, "base url not found")
	}
	baseUrl := os.Getenv("BASE_URL")

	newRequest := c.Request()
	newMultiReader, mrErr := newRequest.MultipartReader()
	if mrErr != nil {
		fmt.Println(mrErr)
		return mrErr
	}

	fileDirName := "temp"

	mkdirErr := os.Mkdir(fileDirName, 0750)
	if mkdirErr != nil {
		if errors.Is(mkdirErr, os.ErrExist) {
			if err := os.RemoveAll(fileDirName); err != nil {
				fmt.Printf("Error deleting temp dir: %v ", err)
			}
			mkdirErr := os.Mkdir(fileDirName, 0750)
			if mkdirErr != nil {
				fmt.Printf("Error creating temp folder: %v", mkdirErr)
			}
		}
	}

	var chatId string
	var uploader []uploadResponse

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

		if part.FileName() != "" {
			fileName := part.FileName()

			caption := fileName
			dst, err := os.Create(fileDirName + "/" + fileName)
			if err != nil {
				fmt.Println("Error creating file:", err)
				return err
			}
			defer dst.Close()

			_, err = io.Copy(dst, part)
			if err != nil {
				fmt.Println("Error copying file:", err)
				return c.JSON(http.StatusGone, err)
			}

			fileStat, _ := dst.Stat()
			fileSize := fileStat.Size()

			if fileSize < 2097152000 {
				sendFile, err := Handlers.SendDocumentRequest(baseUrl, chatId, fileName, fileDirName+"/"+fileName)
				if err != nil {
					return c.JSON(http.StatusGone, err)
				}
        if err := os.Remove(fileDirName+"/"+ fileName);err != nil{fmt.Printf("Error deleting file: %v " , err)}
				json.Unmarshal([]byte(sendFile), &uploader)
				fmt.Printf("upload response: %v", uploader)
			} else {

				dirInfo, err := Handlers.ZipNSplit(dst.Name())
				if err != nil {
					fmt.Println(err)
				}

				for _, file := range dirInfo {
					sendFile, err := Handlers.SendDocumentRequest(baseUrl, chatId, caption, fileDirName+"/"+file.Name())
					if err != nil {
						return c.JSON(http.StatusGone, err)
					}
					var tempResponse uploadResponse
					json.Unmarshal([]byte(sendFile), &tempResponse)
					uploader = append(uploader, tempResponse)
					fmt.Printf("upload response: %v", uploader)
				}
			}
			continue
		}
		if part.FormName() == "url" {
			buf := new(bytes.Buffer)
			buf.ReadFrom(part)
			Url := buf.String()

			client := &http.Client{}
			res, err := client.Get(Url)
			if err != nil {
				fmt.Println(err)
			}
			fileSize, _ := strconv.ParseUint(res.Header.Get("Content-Length"), 10, 64)
			fmt.Printf("file size: %v", fileSize)
			disp := res.Header.Get("Content-Disposition")
			line := strings.Split(disp, "=")
			filename := line[1]
			filename = regexp.MustCompile(`[^a-zA-Z0-9. ]+`).ReplaceAllString(filename, "")
			fmtFileName := strings.ReplaceAll(filename, " ", "")
			curlDlCmd := exec.Command("curl", "-o", fileDirName+"/"+fmtFileName, Url)
			if err := curlDlCmd.Run(); err != nil {
				fmt.Printf("Error downloading %v", err)
			}
			if fileSize < 2097152000 {
				sendFile, err := Handlers.SendDocumentRequest(baseUrl, chatId, line[1], fileDirName+"/"+fmtFileName)
				if err != nil {
					return c.JSON(http.StatusGone, err)
				}

        if err := os.Remove(fileDirName+"/"+ fmtFileName);err != nil{fmt.Printf("Error deleting file: %v " , err)}
				json.Unmarshal([]byte(sendFile), &uploader)
				fmt.Printf("fileId: %v", uploader)
			} else {
				readDir, err := os.ReadDir(fileDirName)
				if err != nil {
					fmt.Printf("Error reading directory %v", err)
				}
				for _, file := range readDir {
					dirInfo, err := Handlers.ZipNSplit(file.Name())
					if err != nil {
						fmt.Println(err)
					}
					caption := file.Name()

					for _, zippedfiles := range dirInfo {
						sendFile, err := Handlers.SendDocumentRequest(baseUrl, chatId, caption, fileDirName+"/"+zippedfiles.Name())
						if err != nil {
							return c.JSON(http.StatusGone, err)
						}
						var tempResponse uploadResponse
						json.Unmarshal([]byte(sendFile), &tempResponse)
						uploader = append(uploader, tempResponse)
						fmt.Printf("upload response: %v", tempResponse)
					}
				}
			}
		}
	}
	return c.JSON(http.StatusOK, uploader)
}
