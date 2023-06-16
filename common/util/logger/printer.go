package logger

import "fmt"

type PrintMode uint8

const (
	PrintModeNormal PrintMode = iota
	PrintModeRaw
)

var mode = PrintModeNormal

func SetPrintMode(value PrintMode) {
	mode = value
}

func Println(message string) {
	if mode == PrintModeNormal {
		fmt.Println(message)
	} else {
		fmt.Printf("%s\r\n", message)
	}
}
func Printf(format string, args ...any) {
	if mode == PrintModeNormal {
		fmt.Printf(format, args...)
	} else {
		fmt.Printf(format+"\r", args...)
	}
}
