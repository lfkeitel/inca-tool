package devices

import "testing"

var testConfig = `
[global]
remote_user = peter
remote_password = cottentail

[boston co-location]
server1
server2 address=10.0.0.2
server3 remote_user=peter1
server4 address=10.0.0.4 protocol=telnet

[san fran location] cisco_enable=orange_cone
server1b
server2b

[web app]
server1
server2b
`

func TestDeviceListParser(t *testing.T) {
	list, err := ParseString(testConfig)
	if err != nil {
		t.Fatal(err)
	}

	if len(list.Groups) != 4 {
		t.Errorf("incorrect number of groups. Expected 4, got %d", len(list.Groups))
	}

	if len(list.Devices) != 6 {
		t.Errorf("incorrect number of devices. Expected 6, got %d", len(list.Devices))
	}

	if list.GetGlobal("remote_user") != "peter" {
		t.Errorf("incorrect global setting remote_user. Expected \"peter\", got \"%s\"", list.GetGlobal("remote_user"))
	}

	if list.GetGlobal("remote_password") != "cottentail" {
		t.Errorf("incorrect global setting remote_password. Expected \"cottentail\", got \"%s\"", list.GetGlobal("remote_password"))
	}

	if list.Devices["server1"].GetSetting("remote_user") != "peter" {
		t.Errorf("incorrect device setting remote_user. Expected \"peter\", got \"%s\"", list.Devices["server1"].GetSetting("remote_user"))
	}

	if list.Devices["server1"].GetSetting("remote_password") != "cottentail" {
		t.Errorf("incorrect device setting remote_password. Expected \"cottentail\", got \"%s\"", list.Devices["server1"].GetSetting("remote_password"))
	}

	if list.Devices["server1"].GetSetting("address") != "" {
		t.Errorf("incorrect device address. Expected \"\", got \"%s\"", list.Devices["server1"].GetSetting("address"))
	}

	if list.Devices["server2"].GetSetting("address") != "10.0.0.2" {
		t.Errorf("incorrect device address. Expected \"10.0.0.2\", got \"%s\"", list.Devices["server2"].GetSetting("address"))
	}

	if list.Devices["server3"].GetSetting("remote_user") != "peter1" {
		t.Errorf("incorrect device setting remote_user. Expected \"peter1\", got \"%s\"", list.Devices["server3"].GetSetting("remote_user"))
	}

	if list.Devices["server4"].GetSetting("protocol") != "telnet" {
		t.Errorf("incorrect device protocol. Expected \"telnet\", got \"%s\"", list.Devices["server4"].GetSetting("protocol"))
	}

	if list.Groups["san fran location"].GetSetting("cisco_enable") != "orange_cone" {
		t.Errorf("incorrect group setting cisco_enable. Expected \"orange_cone\", got \"%s\"", list.Groups["san fran location"].GetSetting("cisco_enable"))
	}

	if len(list.Devices["server1"].Groups) != 2 {
		t.Errorf("incorrect number of group memberships. Expected 2, got %d", len(list.Devices["server1"].Groups))
	}
}
