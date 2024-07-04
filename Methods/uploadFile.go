package Methods

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"tg/telegram-storage/Handlers"
)

func UploadWithCmd(c echo.Context) error {
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

	readTempDir, readDirErr := os.ReadDir(fileDirName)
	if readDirErr != nil {
		if errors.Is(readDirErr, os.ErrNotExist) {
			mkdirErr := os.Mkdir(fileDirName, 0750)
			if mkdirErr != nil {
				fmt.Println(mkdirErr)
			}
		} else {
			if err := os.RemoveAll(fileDirName); err != nil {
				fmt.Println(err)
			}
			mkdirErr := os.Mkdir(fileDirName, 0750)
			if mkdirErr != nil {
				fmt.Println("dir already exist")
			}
		}
	}

	fmt.Printf("red temp dir: %v ", readTempDir)

	var chatId string
	var uploader uploadResponse

	for {
		part, pErr := newMultiReader.NextPart()
		fmt.Printf("here multipart nextpart output: %v\n", part)
		if pErr == io.EOF {
			break
		}

		if part.FormName() == "chat_id" {
			fmt.Println("chatid is running")
			buf := new(bytes.Buffer)
			buf.ReadFrom(part)
			chatId = buf.String()
			continue
		}

		if part.FileName() != "" {
			fmt.Println("file is running")
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

			if err := Handlers.ZipNSplit(dst.Name()); err != nil {
				fmt.Println(err)
			}

			readDir, err := os.ReadDir(fileDirName)
			if err != nil {
				fmt.Printf("Error reading directory %v", err)
			}

			for _, file := range readDir {
				sendFile, err := Handlers.SendDocumentRequest(baseUrl, chatId, caption, fileDirName+"/"+file.Name())
				if err != nil {
					return c.JSON(http.StatusGone, err)
				}
				json.Unmarshal([]byte(sendFile), &uploader)
				fmt.Printf("fileId: %v", uploader)
			}
			continue
		}
		if part.FormName() == "url" {

			fmt.Println("url is running")
			buf := new(bytes.Buffer)
			buf.ReadFrom(part)
			Url := buf.String()

			client := &http.Client{}
			res, err := client.Get(Url)
			if err != nil {
				fmt.Println(err)
			}
			fileSize,_ := strconv.ParseUint(res.Header.Get("Content-Length"), 10 , 64)
			fmt.Printf("heres the file size: %v", fileSize)
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
				sendFile, err := Handlers.SendDocumentRequest(baseUrl, chatId, filename, fileDirName+"/"+fmtFileName)
				if err != nil {
					return c.JSON(http.StatusGone, err)
				}
				json.Unmarshal([]byte(sendFile), &uploader)
				fmt.Printf("fileId: %v", uploader)
			} else {
				readDir, err := os.ReadDir(fileDirName)
				if err != nil {
					fmt.Printf("Error reading directory %v", err)
				}

				for _, file := range readDir {
					if err := Handlers.ZipNSplit(file.Name()); err != nil {
						fmt.Println(err)
					}
					caption := file.Name()
					sendFile, err := Handlers.SendDocumentRequest(baseUrl, chatId, caption, fileDirName+"/"+file.Name())
					if err != nil {
						return c.JSON(http.StatusGone, err)
					}
					json.Unmarshal([]byte(sendFile), &uploader)
					fmt.Printf("fileId: %v", uploader)
				}
			}
		}
	}
	return c.JSON(http.StatusOK, uploader)
}
