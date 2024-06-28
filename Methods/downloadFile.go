package Methods

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"tg/telegram-storage/Handlers"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func DownloadFile(c echo.Context) error {

	type File struct {
		FilePath string `json:"file_path"`
		FileId   string `json:"file_id"`
	}

	type FileResult struct {
		ResponseOk bool `json:"ok"`
		Result     File `json:"result"`
	}

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	request := c.Request()
	request.ParseForm()

	fileIdsArr := request.Form
	baseUrl := os.Getenv("BASE_URL")

	pr, pw := io.Pipe()
	defer pr.Close()

	finalFile, err := os.Create("hello.mkv")
	if err != nil {
		fmt.Println(err)
	}
	defer finalFile.Close()
	

  go func() {
		defer pw.Close()

		filePaths := []string{}
		for _, fileIds := range fileIdsArr {
			for _, fileId := range fileIds {
				fileInfo := Handlers.GetFile(baseUrl, fileId)
				filePaths = append(filePaths, fileInfo.Result.FilePath)
			}
		}

		sort.Strings(filePaths)

		for _, filePath := range filePaths {
			filePart, err := os.Open(filePath)
			fmt.Printf("opened file : %v ", filePart)
			if err != nil {
				fmt.Printf("ERROR OPENING FILE %v ", err)
			}

			stat, err := filePart.Stat()
			if err != nil {
				fmt.Println(err)
				return
			}

			bs := make([]byte, stat.Size())
			_, err = bufio.NewReader(filePart).Read(bs)
			if err != nil && err != io.EOF {
				fmt.Printf("ERROR READING BUFIO %v", err)
				return
			}
			filePart.Close()

			_, writerErr := pw.Write(bs)
			if writerErr != nil {
				fmt.Printf("ERROR WRITING TO SLICE %v ", writerErr)
			}

		}
	}()

	if _, err := io.Copy(finalFile, pr); err != nil {
		fmt.Printf("ERROR COPYING FILE %v", err)
	}

	return c.JSON(http.StatusOK, "success")
}

