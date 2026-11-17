package environment

import "strings"

// Select fields at the CLI as well as in the decoder. Never request daemon
// names, IDs, labels, proxy settings, registry credentials or container lists.
const dockerEndpointFormat = `{{json .Endpoints.docker.Host}}`

const dockerInfoFormat = `{"ServerVersion":{{json .ServerVersion}},"OSType":{{json .OSType}},"Architecture":{{json .Architecture}},"KernelVersion":{{json .KernelVersion}},"CgroupDriver":{{json .CgroupDriver}},"CgroupVersion":{{json .CgroupVersion}},"SecurityOptions":{{json .SecurityOptions}}}`
const dockerVersionFormat = `{"ClientVersion":{{json .Client.Version}},"ClientAPIVersion":{{json .Client.APIVersion}},"ServerVersion":{{if .Server}}{{json .Server.Version}}{{else}}null{{end}},"ServerAPIVersion":{{if .Server}}{{json .Server.APIVersion}}{{else}}null{{end}}}`

type dockerInfo struct {
	ServerVersion   string
	OSType          string
	Architecture    string
	KernelVersion   string
	CgroupDriver    string
	CgroupVersion   string
	SecurityOptions []string
}

type dockerVersion struct {
	ClientVersion    string
	ClientAPIVersion string
	ServerVersion    string
	ServerAPIVersion string
}

func (s *inspection) docker() {
	endpointStatus := s.dockerLocalEndpoint()
	c := s.command("docker", false, "version", "--format", dockerVersionFormat)
	var version dockerVersion
	st := decodeObject(c, &version)
	if st == Pass && (version.ClientVersion == "" || version.ClientAPIVersion == "" || version.ServerVersion == "" || version.ServerAPIVersion == "") {
		st = Fail
	}
	s.add("docker.version", st, true, "Selected Docker client/server versions and API versions only", version)
	c = s.command("docker", false, "info", "--format", dockerInfoFormat)
	var info dockerInfo
	st = decodeObject(c, &info)
	if st == Pass && (info.ServerVersion == "" || info.KernelVersion == "" || info.Architecture == "") {
		st = Fail
	}
	// Do not persist unrecognized security options, which can include paths.
	securityKnown, rootless := len(info.SecurityOptions) > 0, false
	for _, v := range info.SecurityOptions {
		if v == "name=rootless" || v == "rootless" || strings.HasPrefix(v, "name=rootless,") {
			rootless = true
		}
	}
	evidence := struct {
		ServerVersion string `json:"server_version"`
		OS            string `json:"os"`
		Architecture  string `json:"architecture"`
		Kernel        string `json:"kernel"`
		CgroupDriver  string `json:"cgroup_driver"`
		CgroupVersion string `json:"cgroup_version"`
	}{info.ServerVersion, info.OSType, info.Architecture, info.KernelVersion, info.CgroupDriver, info.CgroupVersion}
	s.add("docker.info", st, true, "Selected daemon information; daemon may not be this host", evidence)
	rootStatus, cgroupStatus := st, st
	if st == Pass {
		if !securityKnown {
			rootStatus = Blocked
		} else if rootless {
			rootStatus = Fail
		}
		if info.CgroupVersion != "2" || (info.CgroupDriver != "systemd" && info.CgroupDriver != "cgroupfs") {
			cgroupStatus = Fail
		}
		if info.OSType != "linux" {
			rootStatus, cgroupStatus = Fail, Fail
		}
	}
	s.add("docker.rootful", rootStatus, true, "Require Linux daemon and security-option evidence without rootless mode; empty options are unknown", struct {
		Rootless bool `json:"rootless_reported"`
	}{rootless})
	s.add("docker.cgroup_v2", cgroupStatus, true, "Require daemon cgroup v2 and recognized cgroup driver", nil)

	bindingStatus := endpointStatus
	kernelMatches, archMatches := false, false
	if bindingStatus == Pass {
		bindingStatus = st
		if st == Pass {
			if s.localKernel == "" || s.localArch == "" {
				bindingStatus = Blocked
			} else {
				kernelMatches = info.OSType == "linux" && info.KernelVersion == s.localKernel
				archMatches = dockerArch(info.Architecture) != "" && dockerArch(info.Architecture) == s.GOARCH && dockerArch(info.Architecture) == dockerArch(s.localArch)
				if !kernelMatches || !archMatches {
					bindingStatus = Fail
				}
			}
		}
	}
	s.add("docker.host_consistency", bindingStatus, true, "Require local endpoint and daemon kernel/architecture matching Linux uname and Go runtime; consistency is not cryptographic identity", struct {
		KernelMatches       bool `json:"kernel_matches"`
		ArchitectureMatches bool `json:"architecture_matches"`
	}{kernelMatches, archMatches})
}

// Context inspection alone cannot establish the effective endpoint: Docker
// environment overrides can select a different daemon. Conservatively decline
// host binding for any override rather than copying endpoint/context values
// into arguments or the report, or reimplementing Docker's precedence rules.
func (s *inspection) dockerLocalEndpoint() Status {
	c := s.command("docker", false, "context", "inspect", "--format", dockerEndpointFormat)
	var endpoint string
	st := decodeObject(c, &endpoint)
	standardSocket := st == Pass && endpoint == "unix:///var/run/docker.sock"
	if st == Pass && !standardSocket {
		st = Fail
	}
	envKnown := s.Getenv != nil
	hostOverride, contextOverride := false, false
	if envKnown {
		hostOverride = s.Getenv("DOCKER_HOST") != ""
		contextOverride = s.Getenv("DOCKER_CONTEXT") != ""
	}
	if !envKnown || hostOverride || contextOverride {
		st = NotTested
	}
	if s.GOOS != "linux" {
		st = Blocked
	}
	s.add("docker.local_endpoint", st, true, "Require selected context's standard rootful local Unix endpoint and verified absence of DOCKER_HOST/DOCKER_CONTEXT overrides; raw endpoint omitted", struct {
		EnvironmentChecked     bool `json:"environment_checked"`
		HostOverridePresent    bool `json:"host_override_present"`
		ContextOverridePresent bool `json:"context_override_present"`
		StandardSocketReported bool `json:"standard_socket_reported"`
	}{envKnown, hostOverride, contextOverride, standardSocket})
	return st
}

func dockerArch(value string) string {
	switch value {
	case "amd64", "x86_64":
		return "amd64"
	case "arm64", "aarch64":
		return "arm64"
	default:
		return ""
	}
}
