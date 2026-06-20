package pluginpkg

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildAndValidatePackage(t *testing.T) {
	sourceDir := writeTestPluginSource(t, "1.0.0")
	output := filepath.Join(t.TempDir(), "strm-example.panplugin")
	require.NoError(t, BuildPackage(sourceDir, output))

	validated, err := ValidatePackageFile(context.Background(), output, DefaultPackageLimits())
	require.NoError(t, err)
	assert.Equal(t, "strm-example", validated.Manifest.Name)
	assert.Equal(t, "1.0.0", validated.Manifest.Version)
	assert.NotEmpty(t, validated.Digest)
}

func TestBuildPackageIsReproducible(t *testing.T) {
	sourceDir := writeTestPluginSource(t, "1.0.0")
	first := filepath.Join(t.TempDir(), "first.panplugin")
	second := filepath.Join(t.TempDir(), "second.panplugin")
	require.NoError(t, BuildPackage(sourceDir, first))
	require.NoError(t, BuildPackage(sourceDir, second))
	firstData, err := os.ReadFile(first)
	require.NoError(t, err)
	secondData, err := os.ReadFile(second)
	require.NoError(t, err)
	assert.Equal(t, firstData, secondData)
}

func TestValidatePackageRejectsChecksumMismatch(t *testing.T) {
	sourceDir := writeTestPluginSource(t, "1.0.0")
	output := filepath.Join(t.TempDir(), "strm-example.panplugin")
	require.NoError(t, BuildPackage(sourceDir, output))

	data, err := os.ReadFile(filepath.Join(sourceDir, "manifest.json"))
	require.NoError(t, err)
	data = bytes.Replace(data, []byte("STRM Example"), []byte("Changed Name"), 1)
	require.NoError(t, os.WriteFile(filepath.Join(sourceDir, "manifest.json"), data, 0o600))

	// The existing archive remains valid; directly test the checksum helper with
	// a modified file to ensure tampering is rejected before Wasm execution.
	validated, err := ValidatePackageFile(context.Background(), output, DefaultPackageLimits())
	require.NoError(t, err)
	validated.Files["manifest.json"] = data
	require.Error(t, verifyChecksums(validated.Files))
}

func writeTestPluginSource(t *testing.T, version string) string {
	return writeTestPluginSourceWithWasm(t, version, 1, testPluginWasm())
}

func writeTestPluginSourceWithWasm(
	t *testing.T,
	version string,
	configVersion int,
	wasm []byte,
) string {
	t.Helper()
	dir := t.TempDir()
	manifest := `{
  "schemaVersion": 1,
  "name": "strm-example",
  "version": "` + version + `",
  "displayName": "STRM Example",
  "abi": "pan302-plugin-abi/v2",
  "wasm": "plugin.wasm",
  "configVersion": ` + fmt.Sprint(configVersion) + `,
  "permissions": ["driver.list", "strm.write", "strm.delete", "route.handle"],
  "ui": {
    "manifest": "ui/manifest.json",
    "mode": "trusted-esm"
  }
}
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "manifest.json"), []byte(manifest), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "plugin.wasm"), wasm, 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "default-config.json"), []byte("{}\n"), 0o600))
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "ui"), 0o700))
	uiManifest := `{
  "schemaVersion": 1,
  "name": "strm-example-settings",
  "version": "` + version + `",
  "protocol": "pan302-plugin-ui/v1",
  "mode": "trusted-esm",
  "entry": "index.js",
  "style": "index.css",
  "exportName": "default"
}
`
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ui", "manifest.json"), []byte(uiManifest), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ui", "index.js"), []byte("export default {}\n"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "ui", "index.css"), []byte(".plugin {}\n"), 0o600))
	return dir
}

func testPluginWasmWithMigration(response []byte, initBody []byte) []byte {
	return testPluginWasmWithMigrationAndFree(response, initBody, nil)
}

func testPluginWasmWithMigrationAndFree(
	response []byte,
	initBody []byte,
	freeBody []byte,
) []byte {
	module := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	types := []byte{4}
	types = append(types, functionType([]byte{0x7f, 0x7f, 0x7f, 0x7f}, []byte{0x7f})...)
	types = append(types, functionType([]byte{0x7f}, []byte{0x7f})...)
	types = append(types, functionType([]byte{0x7f, 0x7f}, nil)...)
	types = append(types, functionType([]byte{0x7f, 0x7f}, []byte{0x7e})...)
	module = appendSection(module, 1, types)

	imports := []byte{1}
	imports = appendName(imports, HostModuleName)
	imports = appendName(imports, HostCallName)
	imports = append(imports, 0x00, 0x00)
	module = appendSection(module, 2, imports)

	functions := []byte{6, 1, 2, 3, 3, 3, 3}
	module = appendSection(module, 3, functions)
	module = appendSection(module, 5, []byte{1, 1, 1, 16})

	exports := []byte{7}
	exports = appendExport(exports, "memory", 0x02, 0)
	exports = appendExport(exports, ExportAlloc, 0x00, 1)
	exports = appendExport(exports, ExportFree, 0x00, 2)
	exports = appendExport(exports, ExportInit, 0x00, 3)
	exports = appendExport(exports, ExportOnEvent, 0x00, 4)
	exports = appendExport(exports, ExportHandleHTTP, 0x00, 5)
	exports = appendExport(exports, ExportMigrate, 0x00, 6)
	module = appendSection(module, 7, exports)

	if initBody == nil {
		initBody = []byte{0x00, 0x42, 0x00, 0x0b}
	}
	if freeBody == nil {
		freeBody = []byte{0x00, 0x0b}
	}
	packed := (int64(1024) << 32) | int64(len(response))
	migrateBody := []byte{0x00, 0x42}
	migrateBody = append(migrateBody, encodeI64(packed)...)
	migrateBody = append(migrateBody, 0x0b)
	bodies := [][]byte{
		{0x00, 0x41, 0x08, 0x0b},
		freeBody,
		initBody,
		{0x00, 0x42, 0x00, 0x0b},
		{0x00, 0x42, 0x00, 0x0b},
		migrateBody,
	}
	code := []byte{byte(len(bodies))}
	for _, body := range bodies {
		code = append(code, encodeU32(uint32(len(body)))...)
		code = append(code, body...)
	}
	module = appendSection(module, 10, code)

	data := []byte{1, 0, 0x41}
	data = append(data, encodeI64(1024)...)
	data = append(data, 0x0b)
	data = append(data, encodeU32(uint32(len(response)))...)
	data = append(data, response...)
	return appendSection(module, 11, data)
}

func testPluginWasm() []byte {
	return testPluginWasmWithBodies(
		nil,
		nil,
		nil,
		nil,
		1,
		HostModuleName,
	)
}

func testPluginWasmWithBodies(
	allocBody, initBody, eventBody, httpBody []byte,
	initialMemoryPages uint32,
	importModule string,
) []byte {
	module := []byte{0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00}

	types := []byte{4}
	types = append(types, functionType([]byte{0x7f, 0x7f, 0x7f, 0x7f}, []byte{0x7f})...)
	types = append(types, functionType([]byte{0x7f}, []byte{0x7f})...)
	types = append(types, functionType([]byte{0x7f, 0x7f}, nil)...)
	types = append(types, functionType([]byte{0x7f, 0x7f}, []byte{0x7e})...)
	module = appendSection(module, 1, types)

	imports := []byte{1}
	imports = appendName(imports, importModule)
	imports = appendName(imports, HostCallName)
	imports = append(imports, 0x00, 0x00)
	module = appendSection(module, 2, imports)

	functions := []byte{5, 1, 2, 3, 3, 3}
	module = appendSection(module, 3, functions)
	memory := []byte{1, 1}
	memory = append(memory, encodeU32(initialMemoryPages)...)
	memory = append(memory, encodeU32(max(initialMemoryPages, 16))...)
	module = appendSection(module, 5, memory)

	exports := []byte{6}
	exports = appendExport(exports, "memory", 0x02, 0)
	exports = appendExport(exports, ExportAlloc, 0x00, 1)
	exports = appendExport(exports, ExportFree, 0x00, 2)
	exports = appendExport(exports, ExportInit, 0x00, 3)
	exports = appendExport(exports, ExportOnEvent, 0x00, 4)
	exports = appendExport(exports, ExportHandleHTTP, 0x00, 5)
	module = appendSection(module, 7, exports)

	if allocBody == nil {
		allocBody = []byte{0x00, 0x41, 0x08, 0x0b}
	}
	if initBody == nil {
		initBody = []byte{0x00, 0x42, 0x00, 0x0b}
	}
	if eventBody == nil {
		eventBody = []byte{0x00, 0x42, 0x00, 0x0b}
	}
	if httpBody == nil {
		httpBody = []byte{0x00, 0x42, 0x00, 0x0b}
	}
	bodies := [][]byte{
		allocBody,
		{0x00, 0x0b},
		initBody,
		eventBody,
		httpBody,
	}
	code := []byte{byte(len(bodies))}
	for _, body := range bodies {
		code = append(code, encodeU32(uint32(len(body)))...)
		code = append(code, body...)
	}
	return appendSection(module, 10, code)
}

func functionType(params, results []byte) []byte {
	out := []byte{0x60, byte(len(params))}
	out = append(out, params...)
	out = append(out, byte(len(results)))
	out = append(out, results...)
	return out
}

func appendSection(module []byte, id byte, payload []byte) []byte {
	module = append(module, id)
	module = append(module, encodeU32(uint32(len(payload)))...)
	return append(module, payload...)
}

func appendName(dst []byte, value string) []byte {
	dst = append(dst, encodeU32(uint32(len(value)))...)
	return append(dst, value...)
}

func appendExport(dst []byte, name string, kind byte, index uint32) []byte {
	dst = appendName(dst, name)
	dst = append(dst, kind)
	return append(dst, encodeU32(index)...)
}

func encodeU32(value uint32) []byte {
	var out []byte
	for {
		current := byte(value & 0x7f)
		value >>= 7
		if value != 0 {
			current |= 0x80
		}
		out = append(out, current)
		if value == 0 {
			return out
		}
	}
}

func encodeI64(value int64) []byte {
	var out []byte
	for {
		current := byte(value & 0x7f)
		value >>= 7
		done := (value == 0 && current&0x40 == 0) || (value == -1 && current&0x40 != 0)
		if !done {
			current |= 0x80
		}
		out = append(out, current)
		if done {
			return out
		}
	}
}

var _ = binary.LittleEndian

func max(a, b uint32) uint32 {
	if a > b {
		return a
	}
	return b
}

func TestValidateWasmRejectsMalformedModule(t *testing.T) {
	err := validateWasmModule(
		context.Background(),
		[]byte("not wasm"),
		testRuntimeManifest(),
		DefaultPackageLimits(),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "compile wasm")
}

func TestValidateWasmRejectsUnknownImport(t *testing.T) {
	err := validateWasmModule(
		context.Background(),
		testPluginWasmWithBodies(nil, nil, nil, nil, 1, "wasi_bad1"),
		testRuntimeManifest(),
		DefaultPackageLimits(),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported function import")
}

func TestValidateWasmRejectsMissingRequiredExport(t *testing.T) {
	wasm := bytes.Replace(
		testPluginWasm(),
		[]byte(ExportInit),
		[]byte("pan302_noop"),
		1,
	)
	err := validateWasmModule(
		context.Background(),
		wasm,
		testRuntimeManifest(),
		DefaultPackageLimits(),
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "required export")
}

func TestValidateWasmRejectsInitialMemoryOverLimit(t *testing.T) {
	limits := DefaultPackageLimits()
	wasm := testPluginWasmWithBodies(
		nil,
		nil,
		nil,
		nil,
		limits.MaxInitialMemoryPages+1,
		HostModuleName,
	)
	err := validateWasmModule(context.Background(), wasm, testRuntimeManifest(), limits)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "initial memory exceeds")
}

func testRuntimeManifest() *Manifest {
	manifest := validManifest()
	manifest.Permissions = append(manifest.Permissions, "route.handle")
	return &manifest
}

