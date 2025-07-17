package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
)

type FlagsT struct {
	FlagServerAddr       string
	FlagBaseAddrShortURL string
}

var (
	Flags = FlagsT{}
)

func ParseFlags() (string, string, error) {

	slArg := os.Args[1:]
	if len(slArg) > 2 {
		return "", "", errors.New("количество аргументов командной строки больше двух")
	}

	nameFlagServerAddr := "a"
	nameFlagBaseAddrShortURL := "b"

	for _, v := range slArg {
		rxFlag := v[:2]
		fl := strings.TrimPrefix(rxFlag, "-")
		if fl != nameFlagServerAddr && fl != nameFlagBaseAddrShortURL {
			return "", "", fmt.Errorf("нет поддержки принятого флага {%s}", rxFlag)
		}
	}

	flag.StringVar(&Flags.FlagServerAddr, nameFlagServerAddr, ":8080", "address and port to run server")
	flag.StringVar(&Flags.FlagBaseAddrShortURL, nameFlagBaseAddrShortURL, ":8080/", "base address short URL")
	flag.Parse()

	return Flags.FlagBaseAddrShortURL, Flags.FlagServerAddr, nil
}
