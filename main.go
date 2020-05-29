//go:generate go install -v github.com/josephspurrier/goversioninfo/cmd/goversioninfo
//go:generate goversioninfo -icon=res/papp.ico -manifest=res/papp.manifest
package main

import (
	"os"
	"path"
	"runtime"

	"github.com/portapps/portapps/v2"
	"github.com/portapps/portapps/v2/pkg/log"
	"github.com/portapps/portapps/v2/pkg/registry"
	"github.com/portapps/portapps/v2/pkg/utl"
)

type config struct {
	Cleanup bool `yaml:"cleanup" mapstructure:"cleanup"`
}

var (
	app *portapps.App
	cfg *config
)

func init() {
	var err error

	// Default config
	cfg = &config{
		Cleanup: false,
	}

	// Init app
	if app, err = portapps.NewWithCfg("openvpn-portable", "OpenVPN", cfg); err != nil {
		log.Fatal().Err(err).Msg("Cannot initialize application. See log file for more info.")
	}
}

func main() {
	utl.CreateFolder(app.DataPath)

	appPath := utl.PathJoin(app.AppPath, "win10")
	if app.WinVersion.Major < 10 {
		appPath = utl.PathJoin(app.AppPath, "win7")
	}

	app.Process = utl.PathJoin(appPath, "bin", "openvpn-gui.exe")
	app.WorkingDir = appPath

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
	mainRegKey := registry.Key{
		Key:     `HKLM\SOFTWARE\OpenVPN`,
		Arch:    regArch,
		Default: appPath,
	}
	if err := mainRegKey.Add(true); err != nil {
		log.Error().Err(err).Msg("Cannot add registry key")
	}

	regFile := utl.PathJoin(utl.CreateFolder(app.RootPath, "reg"), "OpenVPN-GUI.reg")
	regKey := registry.Key{
		Key:  `HKCU\Software\OpenVPN-GUI`,
		Arch: "32",
	}

	if err := regKey.Import(regFile); err != nil {
		log.Error().Err(err).Msg("Cannot import registry key")
	}

	defer func() {
		if err := regKey.Export(regFile); err != nil {
			log.Error().Err(err).Msg("Cannot export registry key")
		}
		if cfg.Cleanup {
			if err := mainRegKey.Delete(true); err != nil {
				log.Error().Err(err).Msg("Cannot remove registry key")
			}
			if err := regKey.Delete(true); err != nil {
				log.Error().Err(err).Msg("Cannot remove registry key")
			}
		}
	}()

	defer app.Close()
	app.Launch(os.Args[1:])
}
