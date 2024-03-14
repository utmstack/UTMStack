package utils

import (
	"fmt"
	"sync"

	"github.com/logrusorgru/aurora"
)

var (
	beautyLogger     *BeautyLogger
	beautyLoggerOnce sync.Once
)

type BeautyLogger struct {
}

func GetBeautyLogger() *BeautyLogger {
	beautyLoggerOnce.Do(func() {
		beautyLogger = &BeautyLogger{}
	})
	return beautyLogger
}

func (b *BeautyLogger) WriteError(msg string, err error) {
	if err == nil {
		fmt.Printf("%s: %s: %s\n", "UTMStack", aurora.Red("error").String(), msg)
	} else {
		fmt.Printf("%s: %s: %s: %v\n", "UTMStack", aurora.Red("error").String(), msg, err)
	}
}

func (b *BeautyLogger) WriteSuccessfull(msg string) {
	fmt.Printf("%s: %s: %s\n", "UTMStack", aurora.Green("success").String(), msg)
}

func (b *BeautyLogger) WriteSimpleMessage(msg string) {
	fmt.Println(msg)
}

func (b *BeautyLogger) PrintBanner() {
	banner := "\n" +
		"..........................................................................\n" +
		"  _    _   _                  _____   _                    _  \n" +
		" | |  | | | |                / ____| | |                  | | \n" +
		" | |  | | | |_   _ __ ___   | (___   | |_    __ _    ___  | | __ \n" +
		" | |  | | | __| | '_ ` _ \\   \\___ \\  | __|  / _` |  / __| | |/ / \n" +
		" | |__| | | |_  | | | | | |  ____) | | |_  | (_| | | (__  |   <  \n" +
		"  \\____/   \\__| |_| |_| |_| |_____/   \\__|  \\__,_|  \\___| |_|\\_\\ \n" +
		"..........................................................................\n" +
		"\n"

	fmt.Println(banner)
}
