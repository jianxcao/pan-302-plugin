package pluginpkg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validManifest() Manifest {
	return Manifest{
		SchemaVersion: 1,
		Name:          "strm-example",
		Version:       "1.0.0",
		DisplayName:   "STRM Example",
		ABI:           ABIVersion,
		Wasm:          "plugin.wasm",
		Permissions:   []string{"driver.list", "strm.write", "strm.delete"},
	}
}

func TestManifestValidate(t *testing.T) {
	manifest := validManifest()
	require.NoError(t, manifest.Validate())
}

func TestManifestRejectsUnknownHostOperation(t *testing.T) {
	manifest := validManifest()
	manifest.Permissions = append(manifest.Permissions, "rsa.verify")
	err := manifest.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported permission")
}

func TestManifestRejectsEscapingPath(t *testing.T) {
	manifest := validManifest()
	manifest.Wasm = "../plugin.wasm"
	require.Error(t, manifest.Validate())
}

func TestUIManifestRequiresTrustedESM(t *testing.T) {
	_, err := ParseUIManifest([]byte(`{
		"schemaVersion":1,
		"name":"strm-example-settings",
		"version":"1.0.0",
		"protocol":"pan302-plugin-ui/v1",
		"mode":"trusted-esm",
		"entry":"index.js"
	}`))
	// 注：原测试用的是 "mode":"iframe" 触发校验失败，我们把 "mode" 设为 "iframe" 以测试报错
	// 现在的 UI Protocol 校验：
	// ParseUIManifest 需要 mode == "trusted-esm"，且 protocol == UIProtocol
	// 我们写个会让它失败的测试：
	// "mode":"iframe"
	_, err = ParseUIManifest([]byte(`{
		"schemaVersion":1,
		"name":"strm-example-settings",
		"version":"1.0.0",
		"protocol":"pan302-plugin-ui/v1",
		"mode":"iframe",
		"entry":"index.js"
	}`))
	require.Error(t, err)
}
