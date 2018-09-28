//go:generate go install -v github.com/josephspurrier/goversioninfo/cmd/goversioninfo
//go:generate goversioninfo -icon=res/papp.ico -manifest=res/papp.manifest
package main

import (
	"os"
	"runtime"

	. "github.com/portapps/portapps"
)

func init() {
	Papp.ID = "openvpn-portable"
	Papp.Name = "OpenVPN"
	Init()
}

func main() {
	Papp.AppPath = AppPathJoin("app")
	Papp.DataPath = AppPathJoin("data")
	Papp.Process = PathJoin(Papp.AppPath, "bin", "openvpn-gui.exe")

	configPath := CreateFolder(PathJoin(Papp.DataPath, "config"))
	logPath := CreateFolder(PathJoin(Papp.DataPath, "log"))

	Papp.Args = []string{
		"--exe_path",
		PathJoin(Papp.AppPath, "bin", "openvpn.exe"),
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
	Papp.WorkingDir = Papp.AppPath

	// Add OpenVPN reg key otherwise a dialog popup
	if runtime.GOARCH == "amd64" {
		RegAdd(RegKey{
			Key:  `HKLM\SOFTWARE\OpenVPN`,
			Arch: "64",
		}, true)
	} else {
		RegAdd(RegKey{
			Key:  `HKLM\SOFTWARE\OpenVPN`,
			Arch: "32",
		}, true)
	}

	regsPath := CreateFolder(PathJoin(Papp.Path, "reg"))
	guiRegKey := RegExportImport{
		Key:  `HKCU\Software\OpenVPN-GUI`,
		Arch: "32",
		File: PathJoin(regsPath, "OpenVPN-GUI.reg"),
	}

	ImportRegKey(guiRegKey)
	Launch(os.Args[1:])
	ExportRegKey(guiRegKey)
}
