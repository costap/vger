package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"runtime/debug"

	"github.com/costap/vger/internal/cli/ui"
	"github.com/spf13/cobra"
)

// VersionInfo contains build and runtime version metadata.
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildDate string `json:"build_date"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

var currentVersion VersionInfo

var (
	versionShort bool
	versionJSON  bool
)

// SetVersionInfo sets build metadata injected from package main ldflags
// and supplements missing values from runtime/debug.ReadBuildInfo().
func SetVersionInfo(version, commit, date string) {
	currentVersion = resolveVersionInfo(version, commit, date, debug.ReadBuildInfo)
	Root.Version = currentVersion.Version
	Root.SetVersionTemplate(fmt.Sprintf("vger version %s (%s, %s)\n",
		currentVersion.Version,
		currentVersion.GitCommit,
		currentVersion.BuildDate,
	))
}

// resolveVersionInfo resolves version metadata prioritizing ldflags,
// then runtime/debug build info (module version or VCS revision).
func resolveVersionInfo(v, c, d string, readBuildInfo func() (*debug.BuildInfo, bool)) VersionInfo {
	if readBuildInfo != nil {
		if info, ok := readBuildInfo(); ok {
			// If version wasn't stamped or is generic "dev", check module version
			if (v == "" || v == "dev") && info.Main.Version != "" && info.Main.Version != "(devel)" {
				v = info.Main.Version
			}

			// Extract VCS settings if commit or date were not stamped
			var rev, revTime, modified string
			for _, s := range info.Settings {
				switch s.Key {
				case "vcs.revision":
					rev = s.Value
				case "vcs.time":
					revTime = s.Value
				case "vcs.modified":
					modified = s.Value
				}
			}

			if (c == "" || c == "none" || c == "unknown") && rev != "" {
				if len(rev) > 7 {
					c = rev[:7]
				} else {
					c = rev
				}
				if modified == "true" {
					c += "-dirty"
				}
			}

			if (d == "" || d == "unknown") && revTime != "" {
				d = revTime
			}
		}
	}

	if v == "" {
		v = "dev"
	}
	if c == "" || c == "none" {
		c = "unknown"
	}
	if d == "" {
		d = "unknown"
	}

	return VersionInfo{
		Version:   v,
		GitCommit: c,
		BuildDate: d,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display V'Ger version and build metadata",
	Long:  "Display V'Ger version, git commit, build date, Go runtime version, and platform.",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if currentVersion.Version == "" {
			SetVersionInfo("", "", "")
		}

		if versionJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(currentVersion)
		}

		if versionShort {
			fmt.Println(currentVersion.Version)
			return nil
		}

		ui.VersionCard(
			currentVersion.Version,
			currentVersion.GitCommit,
			currentVersion.BuildDate,
			currentVersion.GoVersion,
			currentVersion.Platform,
		)
		return nil
	},
}

func init() {
	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, "Print only the version string")
	versionCmd.Flags().BoolVar(&versionJSON, "json", false, "Output version information in JSON format")
}
