package viamwpasupplicantmgr

import (
	"fmt"
	"os"
)

const DefaultPremable string = "ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev\nupdate_config=1\n\n"

type Credentials struct {
	SSID string
	PSK string
}


type mgr struct {
	filename string
	preample string
	creds []Credentials
}

// return true if file had to change, false if did nothing
func (m *mgr) checkFileContents() (bool, error) {

	oldBytes, err := os.ReadFile(m.filename)
	if err != nil {
		return false, fmt.Errorf("cannot read filename: %s b/c %w", m.filename, err)
	}

	oldData := string(oldBytes)

	newData := m.contents()

	if newData == oldData {
		return false, nil
	}

	err = os.WriteFile(m.filename, []byte(newData), 0666)
	if err != nil {
		return false, fmt.Errorf("cannot write filename %s b/c %w", m.filename, err)
	}
	
	return true, nil
}

func (m *mgr) contents() string {
	s := m.preample
	if s == "" {
		s = DefaultPremable
	}
	
	for _, c := range m.creds {
		s += fmt.Sprintf("network={\n\tssid=\"%s\"\n\tpsk=\"%s\"\n}\n\n", c.SSID, c.PSK)
	}
	return s
}
