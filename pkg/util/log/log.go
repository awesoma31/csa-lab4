package util

//
// import (
// 	"flag"
// 	"log"
// 	"os"
// )
//
// var (
// 	logLevel string
// 	logger   *log.Logger
// )
//
// func init() {
// 	// Регистрируем флаг для уровня логирования
// 	flag.StringVar(&logLevel, "log-level", "info", "Log level (debug, info, warn, error, none)")
// 	flag.Parse()
//
// 	// Инициализируем логгер
// 	logger = log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
// }
//
// func LogDebug(msg string) {
// 	if logLevel == "debug" {
// 		logger.SetPrefix("DEBUG: ")
// 		logger.Println(msg)
// 	}
// }
//
// func LogInfo(msg string) {
// 	if logLevel == "none" {
// 		return
// 	}
// 	if logLevel == "debug" || logLevel == "info" {
// 		logger.SetPrefix("INFO: ")
// 		logger.Println(msg)
// 	}
// }
//
// func LogWarn(msg string) {
// 	if logLevel == "none" {
// 		return
// 	}
// 	if logLevel != "error" {
// 		logger.SetPrefix("WARN: ")
// 		logger.Println(msg)
// 	}
// }
//
// func LogError(msg string) {
// 	if logLevel == "none" {
// 		return
// 	}
// 	logger.SetPrefix("ERROR: ")
// 	logger.Println(msg)
// }
