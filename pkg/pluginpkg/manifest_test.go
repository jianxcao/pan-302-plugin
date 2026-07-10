package pluginpkg

import (
	"reflect"
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

func TestManifestValidateAllowsMediaEvents(t *testing.T) {
	manifest := validManifest()
	manifest.Events = []string{"media.item.added", "media.item.deleted"}
	manifest.Permissions = append(manifest.Permissions, "event.media.read")

	require.NoError(t, manifest.Validate())
}

func TestManifestContainsOnlyStaticEventDeclarations(t *testing.T) {
	typ := reflect.TypeOf(Manifest{})
	_, exists := typ.FieldByName("EventSourceFilters")
	require.False(t, exists)
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
		"mode":"iframe",
		"entry":"index.js"
	}`))
	require.Error(t, err)
}
