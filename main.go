package main

import (
	//"bytes"
	//"encoding/json"
	//"mime/multipart"
	//"fmt"
	//"io"
	//"os"
	//"log"
	"net/http"
	//"path/filepath"
	//	"time"
	//"fmt"
	//"os"
	//"tg/telegram-storage/Handlers"
	"tg/telegram-storage/Methods"

	//"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
)


func main(){

  //baseUrl:= os.Getenv("BASE_URL")

  app := echo.New()
  app.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "teledisk online!")
	})

  app.POST("/uploadFile" ,Methods.UploadFile)
  app.GET("/downloadFile", Methods.DownloadFile)
  app.Start(":1234")
}
