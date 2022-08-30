package utils

import (
	"github.com/spf13/cobra"
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
	logger, err := loggerConfig.Build()
	encoder := zapcore.NewConsoleEncoder(encodingConfig)

	logger = log.New(
		zapcore.NewCore(encoder, zapcore.AddSync(cmd.OutOrStdout()), zapcore.DebugLevel))

	if err != nil {
		panic(err)
	}

	defer logger.Sync()
	log.ReplaceGlobals(logger)

}
