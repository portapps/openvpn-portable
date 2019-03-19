//go:generate go install -v github.com/josephspurrier/goversioninfo/cmd/goversioninfo
//go:generate goversioninfo -icon=res/papp.ico -manifest=res/papp.manifest
package main

import (
	"os"
	"runtime"

	. "github.com/portapps/portapps"
	"github.com/portapps/portapps/pkg/registry"
	"github.com/portapps/portapps/pkg/utl"
)

var (
	app *App
)

func init() {
	var err error

	// Init app
	if app, err = New("openvpn-portable", "OpenVPN"); err != nil {
		Log.Fatal().Err(err).Msg("Cannot initialize application. See log file for more info.")
	}
}

func main() {
	utl.CreateFolder(app.DataPath)
	app.Process = utl.PathJoin(app.AppPath, "bin", "openvpn-gui.exe")

	configPath := utl.CreateFolder(app.DataPath, "config")
	logPath := utl.CreateFolder(app.DataPath, "log")

	app.Args = []string{
		"--exe_path",
		utl.PathJoin(app.AppPath, "bin", "openvpn.exe"),
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
