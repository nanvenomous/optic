package optic

import (
	"fmt"
)

const (
	colorReset   = "\033[0m"
	colorRed     = "\033[31m"
	colorGreen   = "\033[32m"
	colorYellow  = "\033[33m"
	colorCyan    = "\033[36m"
	colorViolet  = "\033[35m"
	colorBlue    = "\033[34m"
	colorWhite   = "\033[37m"
	colorReverse = "\033[7m"
)

func Optic() string {
	return fmt.Sprintf("%s(o)%s", colorReverse, colorReset)
}

func colorTagPrint(tag string, message string, color string) {
	fmt.Println(fmt.Sprintf("%s %s[%s]%s %s", Optic(), color, tag, colorReset, message))
}

func Green(tag string, message string) {
	colorTagPrint(tag, message, colorGreen)
}

func Cyan(tag string, message string) {
	colorTagPrint(tag, message, colorCyan)
}

func Yellow(tag string, message string) {
	colorTagPrint(tag, message, colorYellow)
}
