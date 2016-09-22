package script

import (
	"io"
	"os"

	"github.com/lfkeitel/inca-tool/src/device"
)

func copyFileContents(src, dst string) error {
	var err error

	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		return err
	}

	defer func() {
		err = out.Close()
	}()

	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}

func getHostVariables(host *device.Device) map[string]string {
	argList := make(map[string]string)
	argList["protocol"] = host.GetSetting("protocol")
	if argList["protocol"] == "" {
		argList["protocol"] = "ssh"
	}

	argList["hostname"] = host.GetSetting("address")
	if argList["hostname"] == "" {
		argList["hostname"] = host.Name
	}

	argList["remote_user"] = host.GetSetting("remote_user")
	if argList["remote_user"] == "" {
		argList["remote_user"] = "root"
	}

	argList["remote_password"] = host.GetSetting("remote_password")

	argList["cisco_enable"] = host.GetSetting("cisco_enable")
	if argList["cisco_enable"] == "" {
		argList["cisco_enable"] = host.GetSetting("remote_password")
	}

	return argList
}
