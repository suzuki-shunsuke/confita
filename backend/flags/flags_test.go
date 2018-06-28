package flags

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/heetch/confita"
	"github.com/stretchr/testify/require"
)

type Config struct {
	A    string        `config:"a"`
	Adef string        `config:"a-def"`
	B    bool          `config:"b"`
	Bdef bool          `config:"b-def"`
	C    time.Duration `config:"c"`
	Cdef time.Duration `config:"c-def"`
	D    int           `config:"d"`
	Ddef int           `config:"d-def"`
}

func runHelper(t *testing.T, args ...string) *Config {
	t.Helper()

	var output bytes.Buffer

	cs := []string{"-test.run=TestHelperProcess", "--"}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Stderr = &output
	cmd.Env = []string{"GO_HELPER_PROCESS=1"}
	err := cmd.Run()
	require.NoError(t, err)

	var cfg Config

	err = json.NewDecoder(&output).Decode(&cfg)
	require.NoError(t, err)

	return &cfg
}

func TestFlags(t *testing.T) {
	t.Run("Use defaults", func(t *testing.T) {
		cfg := runHelper(t, "-a=hello", "-b=true", "-c=10s", "-d=-100")
		require.Equal(t, "hello", cfg.A)
		require.Equal(t, true, cfg.B)
		require.Equal(t, 10*time.Second, cfg.C)
		require.Equal(t, -100, cfg.D)
	})

	t.Run("Override defaults", func(t *testing.T) {
		cfg := runHelper(t, "-a-def=bye", "-b-def=false", "-c-def=15s", "-d-def=-200")
		require.Equal(t, "bye", cfg.Adef)
		require.Equal(t, false, cfg.Bdef)
		require.Equal(t, 15*time.Second, cfg.Cdef)
		require.Equal(t, -200, cfg.Ddef)
	})
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	for len(args) > 0 {
		if args[0] == "--" {
			args = args[1:]
			break
		}
		args = args[1:]
	}
	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No args\n")
		os.Exit(2)
	}

	os.Args = append(os.Args[:1], args...)

	cfg := Config{
		Adef: "hello",
		Bdef: true,
		Cdef: 10 * time.Second,
		Ddef: -100,
	}

	err := confita.NewLoader(NewBackend()).Load(context.Background(), &cfg)
	require.NoError(t, err)
	err = json.NewEncoder(os.Stderr).Encode(&cfg)
	require.NoError(t, err)
	os.Exit(0)
}