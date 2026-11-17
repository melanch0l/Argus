package environment

import (
	"context"
	"testing"
)

func TestFrontendVersionRequirements(t *testing.T) {
	for _, tc := range []struct {
		name, command, output, check string
		want                         Status
	}{
		{"node_baseline", "node --version", "v22.18.0\n", "tool.node", Pass},
		{"node_patch_update", "node --version", "v22.18.1", "tool.node", Pass},
		{"node_minor_update", "node --version", "v22.19.0", "tool.node", Pass},
		{"node_old_minor", "node --version", "v22.17.9", "tool.node", Fail},
		{"node_old_major", "node --version", "v20.19.0", "tool.node", Fail},
		{"node_new_major", "node --version", "v23.0.0", "tool.node", Fail},
		{"node_generic", "node --version", "available", "tool.node", Fail},
		{"node_prerelease", "node --version", "v22.18.0-rc.1", "tool.node", Fail},
		{"node_build_metadata", "node --version", "v22.18.0+custom", "tool.node", Fail},
		{"node_multiline", "node --version", "v22.18.0\ninvalid", "tool.node", Fail},
		{"node_missing_prefix", "node --version", "22.18.0", "tool.node", Fail},
		{"npm_baseline", "npm --version", "10.9.3\n", "tool.npm", Pass},
		{"npm_patch_update", "npm --version", "10.9.4", "tool.npm", Pass},
		{"npm_minor_update", "npm --version", "10.10.0", "tool.npm", Pass},
		{"npm_old_patch", "npm --version", "10.9.2", "tool.npm", Fail},
		{"npm_old_minor", "npm --version", "10.8.9", "tool.npm", Fail},
		{"npm_old_major", "npm --version", "9.9.9", "tool.npm", Fail},
		{"npm_new_major", "npm --version", "11.0.0", "tool.npm", Fail},
		{"npm_generic", "npm --version", "available", "tool.npm", Fail},
		{"npm_prerelease", "npm --version", "10.9.3-rc.1", "tool.npm", Fail},
		{"npm_plus_sign", "npm --version", "+10.9.3", "tool.npm", Fail},
		{"npm_leading_zero", "npm --version", "10.09.3", "tool.npm", Fail},
		{"npm_missing_patch", "npm --version", "10.9", "tool.npm", Fail},
		{"npm_empty_part", "npm --version", "10..3", "tool.npm", Fail},
		{"npm_overflow", "npm --version", "10.999999999999999999999999999.3", "tool.npm", Fail},
	} {
		t.Run(tc.name, func(t *testing.T) {
			i, f, _ := fixture()
			f.results[tc.command] = success(tc.output)
			r := i.Inspect(context.Background(), Options{true})
			assertStatus(t, r, tc.check, tc.want)
			if r.StrictPassed != (tc.want == Pass) {
				t.Fatal("strict gate disagrees with frontend version requirement")
			}
		})
	}
}
