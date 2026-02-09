package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version info (set by ldflags during build)
var (
	Version   = "1.1.0"
	Commit    = "dev"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show mur version",
	RunE:  runVersion,
}

var versionShort bool

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, "Show version only")
}

func runVersion(cmd *cobra.Command, args []string) error {
	if versionShort {
		fmt.Println(Version)
		return nil
	}

	fmt.Printf("mur %s\n", Version)
	fmt.Printf("  commit:  %s\n", Commit)
	fmt.Printf("  built:   %s\n", BuildDate)
	fmt.Printf("  go:      %s\n", runtime.Version())
	fmt.Printf("  os/arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)

	return nil
}
