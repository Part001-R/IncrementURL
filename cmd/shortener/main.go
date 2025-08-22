package main

import (
	"log"

	"github.com/Part001-R/IncrementURL/internal/service"
	"github.com/joho/godotenv"
)

func init() {
	err := godotenv.Load("../../internal/config/.env")
	if err != nil {
		log.Fatalf("ошибка чтения env файла <%v>", err)
	}
}

func main() {
	if err := service.Run(); err != nil {
		log.Fatalf("работа прервана по причине: {%v}", err)
	}
}
