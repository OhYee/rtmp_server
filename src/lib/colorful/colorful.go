package colorful

import "fmt"

// 颜色常量
const (
	Bk = 30
	R  = 31
	G  = 32
	Y  = 33
	Bu = 34
	P  = 35
	S  = 36
	W  = 37

	FrontBlack  = 30
	FrontRed    = 31
	FrontGreen  = 32
	FrontYellow = 33
	FrontBlue   = 34
	FrontPurple = 35
	FrontSky    = 36
	FrontWhite  = 37

	BackBlack  = 40
	BackRed    = 41
	BackGreen  = 42
	BackYellow = 43
	BackBlue   = 44
	BackPurple = 45
	BackSky    = 46
	BackWhite  = 47
)

// FrontBack 颜色显示(前景背景)
func FrontBack(text string, frontColor int, backColor int, args ...interface{}) string {
	return fmt.Sprintf(
		"%s%s\033[0m",
		fmt.Sprintf("\033[1;%d;%dm", frontColor, backColor),
		fmt.Sprintf(text, args...),
	)
}

// Front 颜色显示(前景)
func Front(text string, frontColor int, args ...interface{}) string {
	return fmt.Sprintf(
		"%s%s\033[0m",
		fmt.Sprintf("\033[1;%dm", frontColor),
		fmt.Sprintf(text, args...),
	)
}

// Back 颜色显示(背景)
func Back(text string, backColor int, args ...interface{}) string {
	return fmt.Sprintf(
		"%s%s\033[0m",
		fmt.Sprintf("\033[1;;%dm", backColor),
		fmt.Sprintf(text, args...),
	)
}
