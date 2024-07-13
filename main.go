package main

import (
	"net/http"
	"tg/telegram-storage/Methods"
	"github.com/labstack/echo/v4"
)


func main(){

  app := echo.New()
  app.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "teledisk online!")
	})

  app.POST("/uploadFile" ,Methods.UploadFile)
  app.GET("/downloadFile", Methods.DownloadFile)
  app.Start(":1234")
}
