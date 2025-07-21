package config

import (
	"flag"
)

type FlagsT struct {
	FlagServerAddr       string
	FlagBaseAddrShortURL string
}

var (
	Flags = FlagsT{}
)

func ParseFlags() (string, string, error) {

	var serverAddr string
	var baseAddrShortURL string

	flag.StringVar(&serverAddr, "a", ":8080", "адрес и порт сервера")
	flag.StringVar(&baseAddrShortURL, "b", ":8080/", "базовый адрес для коротких URL")
	flag.Parse()

	/*
		if flag.NArg() > 0 {
			return "", "", errors.New("позиционные аргументы не поддерживаются")
		}
	*/

	return baseAddrShortURL, serverAddr, nil
}
