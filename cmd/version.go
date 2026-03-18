package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Version   = "1.2.2"
	BuildDate = "unknown"
	GitCommit = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print detailed version information including Go version and build details",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ju-cli version: %s\n", Version)
		fmt.Printf("build date: %s\n", BuildDate)
		fmt.Printf("git commit: %s\n", GitCommit)
		fmt.Printf("go version: %s\n", runtime.Version())
		fmt.Printf("platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
