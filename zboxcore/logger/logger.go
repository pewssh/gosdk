package logger

import "github.com/pewssh/gosdk/core/logger"

var defaultLogLevel = logger.DEBUG
var Logger logger.Logger

func init() {
	Logger.Init(defaultLogLevel, "0box-sdk")
}
