package version

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var Version = "development"

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of this CLI",
	RunE: executeVersionCommand,
}

func executeVersionCommand(cmd *cobra.Command, args []string) error {
	log.Infof("{\"version\":\"%v\"}", Version)
	return nil
}