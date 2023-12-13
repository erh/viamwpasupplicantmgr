package viamwpasupplicantmgr

import (
	"context"
	"fmt"
	"os"
	"strings"

	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var MgrModel = resource.ModelNamespace("erh").WithFamily("viamwpasupplicantmgr").WithModel("manager")

const DefaultPremable string = "ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev\nupdate_config=1\n"

func init() {
	resource.RegisterComponent(
		generic.API,
		MgrModel,
		resource.Registration[resource.Resource, *Config]{
			Constructor: newManager,
		})
}

func newManager(ctx context.Context, deps resource.Dependencies, conf resource.Config, logger logging.Logger) (resource.Resource, error) {
	newConf, err := resource.NativeConfig[*Config](conf)
	if err != nil {
		return nil, err
	}

	m := &mgr{
		name: conf.ResourceName(),
		cfg:  newConf,
	}

	_, err = m.checkFileContents()
	if err != nil {
		return nil, err
	}

	return m, nil
}

type Credentials struct {
	SSID    string
	PSK     string
	Encoded bool
}

func (c *Credentials) Equals(other Credentials) bool {
	return c.SSID == other.SSID && c.PSK == other.PSK && c.Encoded == other.Encoded
}

type Config struct {
	resource.TriviallyValidateConfig

	Filename string
	Preample string
	Networks []Credentials
}

type mgr struct {
	resource.TriviallyCloseable
	resource.AlwaysRebuild

	name resource.Name
	cfg  *Config
}

// return true if file had to change, false if did nothing
func (m *mgr) checkFileContents() (bool, error) {

	preample, oldCreds, err := readFile(m.cfg.Filename)
	if err != nil {
		return false, fmt.Errorf("cannot read filename: %s b/c %w", m.cfg.Filename, err)
	}
	if m.cfg.Preample == "" {
		if m.cfg.Preample != "" {
			m.cfg.Preample = preample
		} else {
			m.cfg.Preample = DefaultPremable
		}
	}

	m.cfg.Networks = mergeNetwords(oldCreds, m.cfg.Networks)

	if preample == m.cfg.Preample && networksMatch(oldCreds, m.cfg.Networks) {
		return false, nil
	}

	newData := m.contents()

	err = os.WriteFile(m.cfg.Filename, []byte(newData), 0666)
	if err != nil {
		return false, fmt.Errorf("cannot write filename %s b/c %w", m.cfg.Filename, err)
	}

	return true, nil
}

func (m *mgr) contents() string {
	s := m.cfg.Preample
	if s == "" {
		s = DefaultPremable
	}

	for _, c := range m.cfg.Networks {
		s += "network={\n"
		s += fmt.Sprintf("\tssid=\"%s\"\n", c.SSID)
		if c.Encoded {
			s += fmt.Sprintf("\tpsk=%s\n", c.PSK)
		} else {
			s += fmt.Sprintf("\tpsk=\"%s\"\n", c.PSK)
		}
		s += "}\n"
	}
	return s
}

func (m *mgr) DoCommand(ctx context.Context, cmd map[string]interface{}) (map[string]interface{}, error) {
	didSomething, err := m.checkFileContents()
	return map[string]interface{}{"didSomething": didSomething}, err
}

func (m *mgr) Name() resource.Name {
	return m.name
}

func readFile(filename string) (string, []Credentials, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", nil, err
	}
	return parseFile(string(data))
}

func parseFile(data string) (string, []Credentials, error) {
	lines := strings.Split(data, "\n")

	preample := ""
	creds := []Credentials{}

	inNetwork := false

	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}
		fmt.Printf("line [%s]\n", l)
		if inNetwork {
			if l == "}" {
				inNetwork = false
			} else if strings.HasPrefix(l, "ssid=") {
				l = l[6:]
				l = l[0 : len(l)-1]
				creds[len(creds)-1].SSID = l
			} else if strings.HasPrefix(l, "psk=") {
				l = l[4:]
				if l[0] == '"' {
					l = l[1:]
					l = l[0 : len(l)-1]
				} else {
					creds[len(creds)-1].Encoded = true
				}
				creds[len(creds)-1].PSK = l
			} else {
				return "", nil, fmt.Errorf("in network and bad line [%s]", l)
			}
			continue
		}

		if strings.HasPrefix(l, "network={") {
			inNetwork = true
			creds = append(creds, Credentials{})
		} else {
			preample = preample + l + "\n"
		}
	}
	fmt.Printf("preample [%s]\n", preample)
	return preample, creds, nil
}

func networksMatch(a, b []Credentials) bool {
	if len(a) != len(b) {
		return false
	}

	for _, x := range a {
		idx := findNework(b, x.SSID)
		if idx == -1 {
			return false
		}
		if !x.Equals(b[idx]) {
			return false
		}
	}
	return true
}

func findNework(a []Credentials, ssid string) int {
	for idx, c := range a {
		if c.SSID == ssid {
			return idx
		}
	}
	return -1
}

func mergeNetwords(old, add []Credentials) []Credentials {
	n := []Credentials{}
	for _, c := range old {
		idx := findNework(add, c.SSID)
		if idx == -1 { // it's not in the new list, so add all
			n = append(n, c)
		}
	}
	n = append(n, add...)
	return n
}
