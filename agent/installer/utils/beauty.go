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
		fmt.Printf("%s: %s: %s\n", aurora.Blue("UTMStack").String(), aurora.Red("error").String(), msg)
	} else {
		fmt.Printf("%s: %s: %s: %v\n", aurora.Blue("UTMStack").String(), aurora.Red("error").String(), msg, err)
	}
}

func (b *BeautyLogger) WriteSuccessfull(msg string) {
	fmt.Printf("%s: %s\n", aurora.Green("success").String(), msg)
}

func (b *BeautyLogger) WriteSimpleMessage(msg string) {
	fmt.Println(aurora.Blue(msg).String())
}

func (b *BeautyLogger) PrintBanner() {
	banner := "\n" +
		aurora.Blue("..........................................................................\n").String() +
		aurora.Blue("  _    _   _                  _____   _                    _  \n").String() +
		aurora.Blue(" | |  | | | |                / ____| | |                  | | \n").String() +
		aurora.Blue(" | |  | | | |_   _ __ ___   | (___   | |_    __ _    ___  | | __ \n").String() +
		aurora.Blue(" | |  | | | __| | '_ ` _ \\   \\___ \\  | __|  / _` |  / __| | |/ / \n").String() +
		aurora.Blue(" | |__| | | |_  | | | | | |  ____) | | |_  | (_| | | (__  |   <  \n").String() +
		aurora.Blue("  \\____/   \\__| |_| |_| |_| |_____/   \\__|  \\__,_|  \\___| |_|\\_\\ \n").String() +
		aurora.Blue("..........................................................................\n").String() +
		"\n"

	fmt.Println(banner)
}
