package optic

import (
	"fmt"
	"strings"
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

func optic() string {
	return fmt.Sprintf("%s(o)%s", colorReverse, colorReset)
}

func colorReversePrint(tag string, message string, color string) {
	fmt.Println(fmt.Sprintf("%s%s%s %s %s %s", optic(), colorReverse, color, tag, colorReset, message))
}

func green(tag string, message string) {
	colorReversePrint(tag, message, colorGreen)
}

func cyan(tag string, message string) {
	colorReversePrint(tag, message, colorCyan)
}

func yellow(tag string, message string) {
	colorReversePrint(tag, message, colorYellow)
}

func violet(tag string, message string) {
	colorReversePrint(tag, message, colorViolet)
}

func reflected(code int, path string) {
	var (
		status = fmt.Sprintf("%d", code)
		color  = colorGreen
	)
	if len(status) > 0 {
		switch status[0] {
		case '4':
			color = colorYellow
		case '5':
			color = colorRed
		}
	}

	path = "/" + strings.Replace(path, Base.String(), "", 1)
	colorReversePrint(fmt.Sprintf("%d", code), path, color)
}
