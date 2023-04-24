package utils

import (
	"github.com/spf13/cobra"

	// TODO: we've moved away from logrus to zap; this needs a refactor
	log "go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func ConfigureLoggingForTesting(cmd *cobra.Command) {
	lvl := log.InfoLevel

	loggerConfig := log.NewProductionConfig()
	encodingConfig := log.NewDevelopmentEncoderConfig()
	encodingConfig.TimeKey = ""
	encodingConfig.LevelKey = ""
	encodingConfig.NameKey = ""
	encodingConfig.CallerKey = ""

	loggerConfig.Encoding = "console"
	loggerConfig.Level = log.NewAtomicLevelAt(lvl)
	loggerConfig.EncoderConfig = encodingConfig
	if _, err := loggerConfig.Build(); err != nil {
		panic(err)
	}
	encoder := zapcore.NewConsoleEncoder(encodingConfig)

	logger := log.New(
		zapcore.NewCore(encoder, zapcore.AddSync(cmd.OutOrStdout()), zapcore.DebugLevel))

	defer logger.Sync()
	log.ReplaceGlobals(logger)

}
