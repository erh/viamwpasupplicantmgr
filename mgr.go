package viamwpasupplicantmgr

import (
	"context"
	"fmt"
	"os"

	"go.viam.com/rdk/components/generic"
	"go.viam.com/rdk/logging"
	"go.viam.com/rdk/resource"
)

var MgrModel = resource.ModelNamespace("erh").WithFamily("viamwpasupplicantmgr").WithModel("manager")

const DefaultPremable string = "ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev\nupdate_config=1\n\n"

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

	return &mgr{
		name: conf.ResourceName(),
		cfg:  newConf,
	}, nil
}

type Credentials struct {
	SSID string
	PSK  string
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

	oldBytes, err := os.ReadFile(m.cfg.Filename)
	if err != nil {
		return false, fmt.Errorf("cannot read filename: %s b/c %w", m.cfg.Filename, err)
	}

	oldData := string(oldBytes)

	newData := m.contents()

	if newData == oldData {
		return false, nil
	}

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
		s += fmt.Sprintf("network={\n\tssid=\"%s\"\n\tpsk=\"%s\"\n}\n\n", c.SSID, c.PSK)
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
