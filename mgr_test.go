package viamwpasupplicantmgr

import (
	"os"
	"testing"

	"go.viam.com/test"
)

func TestContents(t *testing.T) {
	m := mgr{
		cfg: &Config{
			Networks: []Credentials{
				{"a", "b", false},
				{"x", "y", false},
				{"c", "d", true},
			},
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
network={
	ssid="c"
	psk=d
}
`,
	)

	pre, creds, err := parseFile(s)
	test.That(t, err, test.ShouldBeNil)
	test.That(t, pre, test.ShouldEqual, DefaultPremable)
	test.That(t, creds, test.ShouldResemble, m.cfg.Networks)

}

func TestCheckFileContents(t *testing.T) {
	f, err := os.CreateTemp("", "viamwpasupplicantmgrtest")
	test.That(t, err, test.ShouldBeNil)
	defer os.Remove(f.Name())

	m := mgr{
		cfg: &Config{
			Filename: f.Name(),
			Networks: []Credentials{
				{"a", "b", true},
				{"x", "y", false},
			},
		},
	}

	didSomething, err := m.checkFileContents()
	test.That(t, err, test.ShouldBeNil)
	test.That(t, didSomething, test.ShouldBeTrue)

	didSomething, err = m.checkFileContents()
	test.That(t, err, test.ShouldBeNil)
	test.That(t, didSomething, test.ShouldBeFalse)
}

func TestHelpers(t *testing.T) {
	a := []Credentials{
		{"a", "b", true},
		{"x", "y", false},
	}
	test.That(t, findNework(a, "a"), test.ShouldEqual, 0)
	test.That(t, findNework(a, "x"), test.ShouldEqual, 1)
	test.That(t, findNework(a, "m"), test.ShouldEqual, -1)

	b := []Credentials{
		{"m", "n", true},
		{"x", "z", false},
	}

	test.That(t, networksMatch(a, []Credentials{}), test.ShouldBeFalse)
	test.That(t, networksMatch(a, b), test.ShouldBeFalse)
	test.That(t, networksMatch(a, a), test.ShouldBeTrue)

	a2 := []Credentials{
		{"a", "c", true},
		{"x", "y", false},
	}
	test.That(t, networksMatch(a, a2), test.ShouldBeFalse)

	m := mergeNetwords(a, b)
	test.That(t, len(m), test.ShouldEqual, 3)
	idx := findNework(m, "x")
	test.That(t, m[idx].PSK, test.ShouldEqual, "z")
}
