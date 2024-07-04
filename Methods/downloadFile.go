package Methods

import (
	"fmt"
	"os/exec"
	"strings"
	"net/http"
	"os"
	"sort"
	"tg/telegram-storage/Handlers"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)

func DownloadFile(c echo.Context) error {

	err := godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	request := c.Request()
	request.ParseForm()

	fileIdsArr := request.Form
	baseUrl := os.Getenv("BASE_URL")

	filePaths := []string{}
	for _, fileIds := range fileIdsArr {
		for _, fileId := range fileIds {
			fileInfo := Handlers.GetFile(baseUrl, fileId)
      rmDefaultFileName, _, _  := strings.Cut(fileInfo.Result.FilePath, "file")
      newFileName := rmDefaultFileName
      if err:= os.Rename(fileInfo.Result.FilePath , newFileName);err != nil{fmt.Println(err)}
			filePaths = append(filePaths, newFileName)
		}
	}

  fmt.Println(filePaths)

	sort.Strings(filePaths)
  finalFilePath := fmt.Sprint(filePaths)
	finalFileName := string(filePaths[0])

  if len(filePaths) > 0{
	joinCmd := exec.Command("cat", finalFilePath, ">", finalFileName + "tar.gz")
	if err := joinCmd.Run(); err != nil {
		fmt.Printf("error joining files %v", err)
		return c.JSON(http.StatusGone, err)
	}

	var tarArgs = "-xvzf"
	tarCmd := exec.Command("tar", tarArgs, finalFileName, ".")
	if err := tarCmd.Run(); err != nil {
		fmt.Printf("error unzipping the file %v", err)
		return c.JSON(http.StatusGone, err)
	}
  }

	return c.JSON(http.StatusOK, "success")
}
