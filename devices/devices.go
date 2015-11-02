package devices

import (
	"fmt"
)

// Device represents a device
type Device struct {
	Name     string
	settings map[string]string
	Groups   []string
	list     *DeviceList
}

// DeviceList is a list of device groups
type DeviceList struct {
	Groups  map[string]*Group
	Devices map[string]*Device
}

// Group is a collection of devices
type Group struct {
	Name     string
	Devices  []*Device
	list     *DeviceList
	settings map[string]string
}

// GetGlobal returns a setting from the global device settings
func (d *DeviceList) GetGlobal(name string) string {
	if _, ok := d.Groups["global"]; !ok {
		return ""
	}
	data, ok := d.Groups["global"].settings[name]
	if !ok {
		return ""
	}
	return data
}

// GetSetting returns the setting name from the group settings. It will also look for global settings if a
// group specific one isn't given. Returns empty string if not found.
func (g *Group) GetSetting(name string) string {
	setting := g.list.GetGlobal(name)
	ns, _ := g.settings[name]
	if ns != "" {
		setting = ns
	}
	return setting
}

// GetSetting returns the setting name from the device's settings.
// The group and global setting will be consulted per the order of precedence.
// Returns empty string if not found.
func (d *Device) GetSetting(name string) string {
	setting := d.list.GetGlobal(name)
	for _, g := range d.Groups {
		ns := d.list.Groups[g].GetSetting(name)
		if ns != "" {
			setting = ns
		}
	}
	ns, _ := d.settings[name]
	if ns != "" {
		setting = ns
	}
	return setting
}

// GetSettings returns all settings as a map from a Group.
func (g *Group) GetSettings() map[string]string {
	return g.settings
}

// GetSettings returns all settings as a map from a Device.
func (d *Device) GetSettings() map[string]string {
	return d.settings
}

// Filter filters a device list to the groups or devices specified
func Filter(dl *DeviceList, filter []string) (*DeviceList, error) {
	devices := &DeviceList{
		Groups:  make(map[string]*Group),
		Devices: make(map[string]*Device),
	}

	for _, term := range filter {
		// Check for a group
		if _, exists := dl.Groups[term]; exists {
			devices.Groups[term] = dl.Groups[term]
			// Add devices from group to Devices field
			for _, d := range devices.Groups[term].Devices {
				devices.Devices[d.Name] = d
			}
			continue
		}
		// Check device
		if _, exists := dl.Devices[term]; exists {
			devices.Devices[term] = dl.Devices[term]
			continue
		}
		return nil, fmt.Errorf("Group or device \"%s\" not found.\n", term)
	}

	return devices, nil
}
