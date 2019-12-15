//go:generate go install -v github.com/josephspurrier/goversioninfo/cmd/goversioninfo
//go:generate goversioninfo -icon=res/papp.ico -manifest=res/papp.manifest
package main

import (
	"os"
	"path"
	"runtime"

	. "github.com/portapps/portapps"
	"github.com/portapps/portapps/pkg/registry"
	"github.com/portapps/portapps/pkg/utl"
)

type config struct {
	Cleanup bool `yaml:"cleanup" mapstructure:"cleanup"`
}

var (
	app *App
	cfg *config
)

func init() {
	var err error

	// Default config
	cfg = &config{
		Cleanup: false,
	}

	// Init app
	if app, err = NewWithCfg("openvpn-portable", "OpenVPN", cfg); err != nil {
		Log.Fatal().Err(err).Msg("Cannot initialize application. See log file for more info.")
	}
}

func main() {
	utl.CreateFolder(app.DataPath)

	appPath := utl.PathJoin(app.AppPath, "win10")
	if WinVersion.Major < 10 {
		appPath = utl.PathJoin(app.AppPath, "win7")
	}

	app.Process = utl.PathJoin(appPath, "bin", "openvpn-gui.exe")

	configPath := utl.CreateFolder(app.DataPath, "config")
	logPath := utl.CreateFolder(app.DataPath, "log")

	app.Args = []string{
		"--exe_path",
		utl.PathJoin(appPath, "bin", "openvpn.exe"),
		"--config_dir",
		configPath,
		"--ext_string",
		"ovpn",
		"--log_dir",
		logPath,
		"--priority_string",
		"NORMAL_PRIORITY_CLASS",
		"--append_string",
		"0",
	}

	// Cleanup on exit
	if cfg.Cleanup {
		defer func() {
			utl.Cleanup([]string{
				path.Join(os.Getenv("USERPROFILE"), "OpenVPN"),
			})
		}()
	}

	// Add OpenVPN reg key otherwise a dialog popup
	regArch := "32"
	if runtime.GOARCH == "amd64" {
		regArch = "64"
	}
	if err := registry.Add(registry.Key{
		Key:  `HKLM\SOFTWARE\OpenVPN`,
		Arch: regArch,
	}, true); err != nil {
		Log.Error().Err(err).Msg("Cannot add registry key")
	}

	regsPath := utl.CreateFolder(app.RootPath, "reg")
	guiRegKey := registry.ExportImport{
		Key:  `HKCU\Software\OpenVPN-GUI`,
		Arch: "32",
		File: utl.PathJoin(regsPath, "OpenVPN-GUI.reg"),
	}

	if err := registry.ImportKey(guiRegKey); err != nil {
		Log.Error().Err(err).Msg("Cannot import registry key")
	}

	defer func() {
		if err := registry.ExportKey(guiRegKey); err != nil {
			Log.Error().Err(err).Msg("Cannot export registry key")
		}
	}()

	app.Launch(os.Args[1:])
}
