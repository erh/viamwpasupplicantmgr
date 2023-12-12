package viamwpasupplicantmgr

import (
	"os"
	"testing"

	"go.viam.com/test"
)

func TestContents(t *testing.T) {
	m := mgr{
		creds: []Credentials{
			{"a", "b"},
			{"x", "y"},
		},
	}


	s := m.contents()
	test.That(t, s, test.ShouldEqual, `ctrl_interface=DIR=/var/run/wpa_supplicant GROUP=netdev
update_config=1

network={
	ssid="a"
	psk="b"
}

network={
	ssid="x"
	psk="y"
}

`,
	)
}

func TestCheckFileContents(t *testing.T) {
	f, err := os.CreateTemp("", "viamwpasupplicantmgrtest")
	test.That(t, err, test.ShouldBeNil)
	defer os.Remove(f.Name())

	m := mgr{
		filename: f.Name(),
		creds: []Credentials{
			{"a", "b"},
			{"x", "y"},
		},
	}

	didSomething, err := m.checkFileContents()
	test.That(t, err, test.ShouldBeNil)
	test.That(t, didSomething, test.ShouldBeTrue)

	didSomething, err = m.checkFileContents()
	test.That(t, err, test.ShouldBeNil)
	test.That(t, didSomething, test.ShouldBeFalse)
}
