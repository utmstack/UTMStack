package utils

import (
	"fmt"
	"sync"

	"github.com/logrusorgru/aurora"
	"github.com/threatwinds/logger"
)

var (
	Logger           *BeautyLogger
	beautyLoggerOnce sync.Once
)

type BeautyLogger struct {
	name       string
	fileLogger *logger.Logger
}

func InitBeautyLogger(name, filepath string) {
	beautyLoggerOnce.Do(func() {
		Logger = &BeautyLogger{}
		Logger.name = name
		Logger.fileLogger = logger.NewLogger(
			&logger.Config{Format: "text", Level: 200, Output: filepath, Retries: 3, Wait: 5},
		)
	})
}

func (b *BeautyLogger) WriteError(format string, args ...any) {
	fmt.Printf(fmt.Sprintf("%s: %s: %s\n", b.name, aurora.Red("error").String(), format), args...)
	b.fileLogger.ErrorF(fmt.Sprintf("%s: %s", b.name, format), args...)
}

func (b *BeautyLogger) WriteFatal(format string, args ...any) {
	fmt.Printf(fmt.Sprintf("%s: %s: %s\n", b.name, aurora.Red("error").String(), format), args...)
	b.fileLogger.Fatal(fmt.Sprintf("%s: %s", b.name, format), args...)
}

func (b *BeautyLogger) WriteSimpleMessage(format string, args ...any) {
	fmt.Printf(fmt.Sprintf("%s: %s\n", b.name, format), args...)
	b.fileLogger.Info(fmt.Sprintf("%s: %s", b.name, format), args...)
}

func (b *BeautyLogger) WriteSuccessfull(format string, args ...any) {
	fmt.Printf(fmt.Sprintf("%s: %s: %s\n", b.name, aurora.Green("success").String(), format), args...)
	b.fileLogger.Info(fmt.Sprintf("%s: %s", b.name, format), args...)
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
