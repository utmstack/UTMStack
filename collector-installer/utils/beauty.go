package utils

import (
	"fmt"
	"sync"

	"github.com/threatwinds/logger"

	"github.com/logrusorgru/aurora"
)

var (
	beautyLogger     *BeautyLogger
	beautyLoggerOnce sync.Once
)

type BeautyLogger struct {
	fileLogger *logger.Logger
}

func GetBeautyLogger(filepath string) *BeautyLogger {
	beautyLoggerOnce.Do(func() {
		beautyLogger = &BeautyLogger{}
		beautyLogger.fileLogger = CreateLogger(filepath)
	})
	return beautyLogger
}

func (b *BeautyLogger) WriteError(msg string, err error) {
	if err == nil {
		fmt.Printf("%s: %s: %s\n", "UTMStack Collector", aurora.Red("error").String(), msg)
		b.fileLogger.ErrorF(fmt.Sprintf("%s: %s: %s", "UTMStack Collector", "error", msg))
	} else {
		fmt.Printf("%s: %s: %s: %v\n", "UTMStack Collector", aurora.Red("error").String(), msg, err)
		b.fileLogger.ErrorF(fmt.Sprintf("%s: %s: %s: %v", "UTMStack Collector", "error", msg, err))
	}
}

func (b *BeautyLogger) WriteFatal(msg string, err error) {
	if err == nil {
		fmt.Printf("%s: %s: %s\n", "UTMStack Collector", aurora.Red("error").String(), msg)
		b.fileLogger.Fatal(fmt.Sprintf("%s: %s: %s", "UTMStack Collector", "error", msg))
	} else {
		fmt.Printf("%s: %s: %s: %v\n", "UTMStack Collector", aurora.Red("error").String(), msg, err)
		b.fileLogger.Fatal(fmt.Sprintf("%s: %s: %s: %v", "UTMStack Collector", "error", msg, err))
	}
}

func (b *BeautyLogger) WriteSuccessfull(msg string) {
	fmt.Printf("%s: %s: %s\n", "UTMStack Collector", aurora.Green("success").String(), msg)
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
