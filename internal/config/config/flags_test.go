package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ParseFlags_SUCCESS(t *testing.T) {

	testData := []struct {
		testName string
		arg_cmd  string
		arg_a    string
		arg_b    string
		wantAddr string
		wantBase string
	}{
		{
			testName: "correct data",
			arg_cmd:  "cmd",
			arg_a:    "-a=localhost:9999",
			arg_b:    "-b=http://localhost:5500/",
			wantAddr: "localhost:9999",
			wantBase: "http://localhost:5500/",
		},
	}

	for _, tt := range testData {
		t.Run(tt.testName, func(t *testing.T) {

			os.Args = []string{tt.arg_cmd, tt.arg_a, tt.arg_b}
			ParseFlags()

			assert.Equalf(t, tt.wantAddr, Flags.FlagServerAddr, "ожидалось {%s}, а принято {%s}", tt.wantAddr, Flags.FlagServerAddr)
			assert.Equalf(t, tt.wantBase, Flags.FlagBaseAddrShortURL, "ожидалось {%s}, а принято {%s}", tt.wantAddr, tt.wantBase, Flags.FlagBaseAddrShortURL)
		})
	}
}

// Реакция на неподдерживаемое имя флага
func Test_ParseFlags_FAULT_1(t *testing.T) {

	testData := []struct {
		testName  string
		arg_cmd   string
		arg_a     string
		arg_b     string
		wantError string
	}{
		{
			testName:  "wrong name arg_a",
			arg_cmd:   "cmd",
			arg_a:     "-c=localhost:9999",
			arg_b:     "-b=http://localhost:5500/",
			wantError: "нет поддержки принятого флага {-c}",
		},
	}

	for _, tt := range testData {
		t.Run(tt.testName, func(t *testing.T) {

			os.Args = []string{tt.arg_cmd, tt.arg_a, tt.arg_b}

			err := ParseFlags()
			assert.Equalf(t, tt.wantError, err.Error(), "ожидалось {%s}, а принято {%s}", tt.wantError, err.Error())
		})
	}
}

// Реакция на избыточное количесвтво флагов
func Test_ParseFlags_FAULT_2(t *testing.T) {

	testData := []struct {
		testName  string
		arg_cmd   string
		arg_a     string
		arg_b     string
		wantError string
	}{
		{
			testName:  "wrong name arg_a",
			arg_cmd:   "cmd",
			arg_a:     "-a=localhost:9999",
			arg_b:     "-b=http://localhost:5500/",
			wantError: "количество аргументов командной строки больше двух",
		},
	}

	for _, tt := range testData {
		t.Run(tt.testName, func(t *testing.T) {

			os.Args = []string{tt.arg_cmd, tt.arg_a, tt.arg_b, tt.arg_b}

			err := ParseFlags()
			assert.Equalf(t, tt.wantError, err.Error(), "ожидалось {%s}, а принято {%s}", tt.wantError, err.Error())
		})
	}
}
