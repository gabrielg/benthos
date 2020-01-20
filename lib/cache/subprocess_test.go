package cache

import (
	"os"
	"os/exec"
	"testing"

	"github.com/Jeffail/benthos/v3/lib/log"
	"github.com/Jeffail/benthos/v3/lib/metrics"
	"github.com/Jeffail/benthos/v3/lib/types"
)

//------------------------------------------------------------------------------

func TestSubprocessCacheGet(t *testing.T) {
	testLog := log.New(os.Stdout, log.Config{LogLevel: "NONE"})

	conf := NewConfig()
	conf.Type = "subprocess"
	conf.Subprocess.Name = "sh"
	conf.Subprocess.Args = []string{
		"-c",
		`[ "$2" = "testkey" ] && printf testval`,
		"--",
	}

	c, err := New(conf, nil, testLog, metrics.DudType{})
	if err != nil {
		t.Fatal(err)
	}

	expErr := types.ErrKeyNotFound
	if _, act := c.Get("missingkey"); act != expErr {
		t.Errorf("Wrong error returned: %v != %v", act, expErr)
	}

	exp := "testval"
	if act, err := c.Get("testkey"); err != nil {
		t.Error(err)
	} else if string(act) != exp {
		t.Errorf("Wrong result: %v != %v", string(act), exp)
	}
}

func TestSubprocessCacheGetWithError(t *testing.T) {
	testLog := log.New(os.Stdout, log.Config{LogLevel: "NONE"})

	conf := NewConfig()
	conf.Type = "subprocess"
	conf.Subprocess.Name = "sh"
	conf.Subprocess.Args = []string{
		"-c",
		`printf "an error occurred" >&2 && exit 2`,
	}

	c, err := New(conf, nil, testLog, metrics.DudType{})
	if err != nil {
		t.Fatal(err)
	}

	val, act := c.Get("boom")

	exp := "an error occurred"
	switch act.(type) {
	case nil:
		t.Errorf("Expected an error but no error returned: %v", val)
	case *exec.ExitError:
		if string(val) != exp {
			t.Errorf("Wrong result: %v != %v", string(val), exp)
		}
	default:
		t.Errorf("Wrong error returned: %v", act)
	}
}
