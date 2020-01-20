package cache

import (
	"bytes"
	"os/exec"
	"time"

	"github.com/Jeffail/benthos/v3/lib/log"
	"github.com/Jeffail/benthos/v3/lib/metrics"
	"github.com/Jeffail/benthos/v3/lib/types"
)

//------------------------------------------------------------------------------

func init() {
	Constructors[TypeSubprocess] = TypeSpec{
		constructor: NewSubprocess,
		Description: `
The subprocess cache executes another process to get values from a cache. At the
moment it only supports ` + "`get`" + `operations. If the process exits with an
exit value of ` + "`0`" + ` the value returned over ` + "`stdout`" + ` is used
as the cached value. An exit status of ` + "`1`" + `indicates that the key was
not found in the cache. All other exit statuses are considered an unhandled
error state.

The given command will be excuted with the operation (e.g. ` + "`get`" + `) as
the first argument, and the key as the second argument.

The field ` + "`args`" + ` can be used to provide extra arguments that should
be passed to the command, they will be inserted before the operation and key.

` + "```yaml" + `
type: subprocess
subprocess:
  name: sh
  args:
	- "-c"
	- "printf $1 $2"
	- "--"
` + "```" + `

These values can be overridden during execution, at which point the configured
TTL is respected as usual.`,
	}
}

//------------------------------------------------------------------------------

// SubprocessConfig contains config fields for the Subprocess cache type.
type SubprocessConfig struct {
	Name string   `json:"name" yaml:"name"`
	Args []string `json:"args" yaml:"args"`
}

// NewSubprocessConfig creates a SubprocessConfig populated with default values.
func NewSubprocessConfig() SubprocessConfig {
	return SubprocessConfig{
		Name: "",
		Args: []string{},
	}
}

//------------------------------------------------------------------------------

// Subprocess is an external executable based cache implementation.
type Subprocess struct {
	name string
	args []string
}

// NewSubprocess creates a new Subprocess cache type.
func NewSubprocess(conf Config, mgr types.Manager, log log.Modular, stats metrics.Type) (types.Cache, error) {
	return &Subprocess{conf.Subprocess.Name, conf.Subprocess.Args}, nil
}

const (
	exitCodeKeyNotFound = 1
)

//------------------------------------------------------------------------------

func (m *Subprocess) run(operation string, args ...string) ([]byte, error) {
	allArgs := append(m.args, operation)
	allArgs = append(allArgs, args...)

	cmd := exec.Command(m.name, allArgs...)
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	value, err := cmd.Output()

	if err != nil {
		return stderr.Bytes(), err
	}
	return value, nil
}

// Get attempts to locate and return a cached value by its key, returns an error
// if the key does not exist. A key's non-existence is signalled by an exit code
// of `1` in order to differentiate from an intentional cached empty value. Any
// other exit code is considered an unknown error state.
func (m *Subprocess) Get(key string) ([]byte, error) {
	value, err := m.run("get", key)

	switch e := err.(type) {
	case nil:
		return value, nil
	case *exec.ExitError:
		if e.ExitCode() == exitCodeKeyNotFound {
			return nil, types.ErrKeyNotFound
		}
		return value, err
	default:
		return value, err
	}
}

// Set attempts to set the value of a key.
func (m *Subprocess) Set(key string, value []byte) error {
	return nil
}

// SetMulti attempts to set the value of multiple keys, returns an error if any
// keys fail.
func (m *Subprocess) SetMulti(items map[string][]byte) error {
	return nil
}

// Add attempts to set the value of a key only if the key does not already exist
// and returns an error if the key already exists.
func (m *Subprocess) Add(key string, value []byte) error {
	return nil
}

// Delete attempts to remove a key.
func (m *Subprocess) Delete(key string) error {
	return nil
}

// CloseAsync shuts down the cache.
func (m *Subprocess) CloseAsync() {
}

// WaitForClose blocks until the cache has closed down.
func (m *Subprocess) WaitForClose(timeout time.Duration) error {
	return nil
}

//------------------------------------------------------------------------------
