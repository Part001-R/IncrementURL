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

func ParseFlags() error {

	slArg := os.Args[1:]
	if len(slArg) > 2 {
		return errors.New("количество аргументов командной строки больше двух")
	}

	name_FlagServerAddr := "a"
	name_FlagBaseAddrShortURL := "b"

	for _, v := range slArg {
		rxFlag := v[:2]
		fl := strings.TrimPrefix(rxFlag, "-")
		if fl != name_FlagServerAddr && fl != name_FlagBaseAddrShortURL {
			return fmt.Errorf("нет поддержки принятого флага {%s}", rxFlag)
		}
	}

	flag.StringVar(&Flags.FlagServerAddr, name_FlagServerAddr, ":8080", "address and port to run server")
	flag.StringVar(&Flags.FlagBaseAddrShortURL, name_FlagBaseAddrShortURL, ":8080/", "base address short URL")
	flag.Parse()

	return nil
}
