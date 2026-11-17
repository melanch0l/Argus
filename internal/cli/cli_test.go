package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/melanch0l/Argus/internal/buildinfo"
)

func TestCommands(t *testing.T) {
	for _, component := range []Component{Agent, Backend} {
		t.Run(string(component), func(t *testing.T) {
			cases := []struct {
				name   string
				args   []string
				code   int
				want   string
				stdout bool
			}{
				{"default", nil, ExitNotImplemented, "not_implemented", false},
				{"version", []string{"--version"}, ExitOK, buildinfo.Version + " " + buildinfo.Phase, true},
				{"help", []string{"--help"}, ExitOK, "Usage:", true},
				{"short help", []string{"-h"}, ExitOK, "Usage:", true},
				{"unknown", []string{"--canary-secret"}, ExitUsage, "invalid command line", false},
				{"bad bool", []string{"--check-config=canary-secret"}, ExitUsage, "invalid command line", false},
				{"missing config value", []string{"--config"}, ExitUsage, "invalid command line", false},
				{"positional", []string{"canary-secret"}, ExitUsage, "invalid flag combination", false},
				{"check missing path", []string{"--check-config"}, ExitUsage, "invalid flag combination", false},
				{"config alone", []string{"--config", "canary-secret"}, ExitUsage, "invalid flag combination", false},
				{"version check", []string{"--version", "--check-config"}, ExitUsage, "invalid flag combination", false},
				{"version config", []string{"--version", "--config", "canary-secret"}, ExitUsage, "invalid flag combination", false},
				{"missing file", []string{"--check-config", "--config", filepath.Join(t.TempDir(), "canary-secret")}, ExitUsage, "cannot open configuration file", false},
			}
			for _, tt := range cases {
				t.Run(tt.name, func(t *testing.T) {
					var out, errs bytes.Buffer
					if code := Run(component, tt.args, &out, &errs); code != tt.code {
						t.Fatalf("exit %d, want %d", code, tt.code)
					}
					got := errs.String()
					if tt.stdout {
						got = out.String()
						if errs.Len() != 0 {
							t.Fatal("unexpected stderr")
						}
					} else if out.Len() != 0 {
						t.Fatal("unexpected stdout")
					}
					if !strings.Contains(got, tt.want) {
						t.Fatalf("missing %q in %q", tt.want, got)
					}
					if strings.Contains(out.String()+errs.String(), "canary-secret") {
						t.Fatal("echoed caller data")
					}
					if tt.name == "version" && out.String() != "argus-"+string(component)+" "+buildinfo.Version+" "+buildinfo.Phase+"\n" {
						t.Fatal("unexpected version output")
					}
				})
			}
		})
	}
}

func TestCheckConfigNoSecretOrStateAccess(t *testing.T) {
	for _, component := range []Component{Agent, Backend} {
		t.Run(string(component), func(t *testing.T) {
			// None of these deployment paths exist. A syntax/settings-only check must
			// still pass, without creating state or trying to read credentials.
			dir := t.TempDir()
			token := filepath.Join(dir, "missing-token")
			state := filepath.Join(dir, "missing-state")
			socket := filepath.Join(dir, "missing-socket")
			password := filepath.Join(dir, "missing-password")
			text := `{"version":1,"token_file":"` + token + `",`
			if component == Agent {
				text += `"backend_url":"https://backend.example.invalid/api/v1","state_dir":"` + state + `","docker_socket":"` + socket + `"}`
			} else {
				text += `"postgresql":{"host":"postgres.example.invalid","database":"argus","user":"argus_app","password_file":"` + password + `"}}`
			}
			filename := filepath.Join(dir, "config.json")
			if err := os.WriteFile(filename, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("ARGUS_TOKEN", "canary-secret")
			t.Setenv("DATABASE_URL", "canary-secret")
			var out, errs bytes.Buffer
			code := Run(component, []string{"--check-config", "--config", filename}, &out, &errs)
			if code != ExitOK || errs.Len() != 0 || !strings.Contains(out.String(), "syntax/settings only") {
				t.Fatalf("check failed: %d %s %s", code, &out, &errs)
			}
			for _, p := range []string{token, state, socket, password} {
				if _, err := os.Stat(p); !os.IsNotExist(err) {
					t.Fatal("check created deployment path")
				}
			}
			entries, err := os.ReadDir(dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(entries) != 1 {
				t.Fatal("unexpected side effects")
			}
		})
	}
}

func TestConfigFailuresSafe(t *testing.T) {
	for _, component := range []Component{Agent, Backend} {
		for _, text := range []string{
			`{"canary-secret":"canary-secret"}`, `{"version":"canary-secret"}`, `{"version":1,"version":1}`, `{} {}`, `{"canary-secret":`,
		} {
			filename := filepath.Join(t.TempDir(), "canary-secret.json")
			if err := os.WriteFile(filename, []byte(text), 0600); err != nil {
				t.Fatal(err)
			}
			var out, errs bytes.Buffer
			if code := Run(component, []string{"--check-config", "--config", filename}, &out, &errs); code != ExitUsage {
				t.Fatal("accepted invalid config")
			}
			if out.Len() != 0 || !strings.Contains(errs.String(), "configuration invalid:") || strings.Contains(errs.String(), "canary-secret") {
				t.Fatal("unsafe error output")
			}
		}
	}
}

func TestInvalidComponent(t *testing.T) {
	var out, errs bytes.Buffer
	if Run("canary-secret", nil, &out, &errs) != ExitUsage || strings.Contains(errs.String(), "canary-secret") {
		t.Fatal("unsafe component handling")
	}
}
