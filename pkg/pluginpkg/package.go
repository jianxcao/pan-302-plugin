package pluginpkg

import (
	"archive/zip"
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

const (
	// 插件 ZIP 不允许超过 128MiB，避免下载/传输成本；运行时内存已设 256MiB，足以容纳
	// 插件执行所需数据。实际解压大小受 RuntimeOptions.MemoryLimitPages 限制。
	DefaultMaxPackageSize   int64 = 128 << 20
	DefaultMaxExtractedSize int64 = 256 << 20
	DefaultMaxFileSize      int64 = 64 << 20
	DefaultMaxFiles               = 1024
	// 允许插件声明最多 128MiB 初始内存；运行期仍由 RuntimeOptions.MemoryLimitPages
	// 限制总增长，目前默认 256MiB。
	DefaultMaxInitialMemoryPages = 2048
)

var deterministicZipTime = time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)

var allowedWASIImports = map[string]struct{}{
	"args_get":                {},
	"args_sizes_get":          {},
	"clock_res_get":           {},
	"clock_time_get":          {},
	"environ_get":             {},
	"environ_sizes_get":       {},
	"fd_advise":               {},
	"fd_allocate":             {},
	"fd_close":                {},
	"fd_datasync":             {},
	"fd_fdstat_get":           {},
	"fd_fdstat_set_flags":     {},
	"fd_fdstat_set_rights":    {},
	"fd_filestat_get":         {},
	"fd_filestat_set_size":    {},
	"fd_filestat_set_times":   {},
	"fd_pread":                {},
	"fd_prestat_dir_name":     {},
	"fd_prestat_get":          {},
	"fd_pwrite":               {},
	"fd_read":                 {},
	"fd_readdir":              {},
	"fd_renumber":             {},
	"fd_seek":                 {},
	"fd_sync":                 {},
	"fd_tell":                 {},
	"fd_write":                {},
	"path_create_directory":   {},
	"path_filestat_get":       {},
	"path_filestat_set_times": {},
	"path_link":               {},
	"path_open":               {},
	"path_readlink":           {},
	"path_remove_directory":   {},
	"path_rename":             {},
	"path_symlink":            {},
	"path_unlink_file":        {},
	"poll_oneoff":             {},
	"proc_exit":               {},
	"proc_raise":              {},
	"random_get":              {},
	"sched_yield":             {},
	"sock_accept":             {},
	"sock_recv":               {},
	"sock_send":               {},
	"sock_shutdown":           {},
}

type PackageLimits struct {
	MaxPackageSize        int64
	MaxExtractedSize      int64
	MaxFileSize           int64
	MaxFiles              int
	MaxInitialMemoryPages uint32
	MaxVersions           int // 最多保留的版本数（含 active），0 表示不限制
}

func DefaultPackageLimits() PackageLimits {
	return PackageLimits{
		MaxPackageSize:        DefaultMaxPackageSize,
		MaxExtractedSize:      DefaultMaxExtractedSize,
		MaxFileSize:           DefaultMaxFileSize,
		MaxFiles:              DefaultMaxFiles,
		MaxInitialMemoryPages: DefaultMaxInitialMemoryPages,
		MaxVersions:           2, // 默认保留 active + 1 个回滚版本
	}
}

type ValidatedPackage struct {
	Manifest *Manifest
	Files    map[string][]byte
	Digest   string
}

func ValidatePackageFile(ctx context.Context, packagePath string, limits PackageLimits) (*ValidatedPackage, error) {
	info, err := os.Stat(packagePath)
	if err != nil {
		return nil, err
	}
	if info.Size() > limits.MaxPackageSize {
		return nil, fmt.Errorf("%w: package exceeds %d bytes", ErrInvalidPackage, limits.MaxPackageSize)
	}
	reader, err := zip.OpenReader(packagePath)
	if err != nil {
		return nil, fmt.Errorf("%w: open ZIP: %v", ErrInvalidPackage, err)
	}
	defer reader.Close()
	return validateZip(ctx, &reader.Reader, limits)
}

func validateZip(ctx context.Context, reader *zip.Reader, limits PackageLimits) (*ValidatedPackage, error) {
	if len(reader.File) == 0 || len(reader.File) > limits.MaxFiles {
		return nil, fmt.Errorf("%w: invalid file count %d", ErrInvalidPackage, len(reader.File))
	}
	files := make(map[string][]byte, len(reader.File))
	caseFolded := make(map[string]string, len(reader.File))
	var total int64
	for _, file := range reader.File {
		name := file.Name
		if err := validatePackagePath(name); err != nil {
			return nil, fmt.Errorf("%w: %q: %v", ErrInvalidPackage, name, err)
		}
		if file.FileInfo().IsDir() {
			continue
		}
		if !file.Mode().IsRegular() {
			return nil, fmt.Errorf("%w: %q is not a regular file", ErrInvalidPackage, name)
		}
		folded := strings.ToLower(name)
		if previous, ok := caseFolded[folded]; ok {
			return nil, fmt.Errorf("%w: conflicting paths %q and %q", ErrInvalidPackage, previous, name)
		}
		caseFolded[folded] = name
		if _, exists := files[name]; exists {
			return nil, fmt.Errorf("%w: duplicate path %q", ErrInvalidPackage, name)
		}
		if file.UncompressedSize64 > uint64(limits.MaxFileSize) {
			return nil, fmt.Errorf("%w: %q exceeds file size limit", ErrInvalidPackage, name)
		}
		total += int64(file.UncompressedSize64)
		if total > limits.MaxExtractedSize {
			return nil, fmt.Errorf("%w: extracted content exceeds size limit", ErrInvalidPackage)
		}
		rc, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("%w: open %q: %v", ErrInvalidPackage, name, err)
		}
		data, readErr := io.ReadAll(io.LimitReader(rc, limits.MaxFileSize+1))
		closeErr := rc.Close()
		if readErr != nil {
			return nil, fmt.Errorf("%w: read %q: %v", ErrInvalidPackage, name, readErr)
		}
		if closeErr != nil {
			return nil, fmt.Errorf("%w: close %q: %v", ErrInvalidPackage, name, closeErr)
		}
		if int64(len(data)) > limits.MaxFileSize {
			return nil, fmt.Errorf("%w: %q exceeds file size limit", ErrInvalidPackage, name)
		}
		files[name] = data
	}

	manifestData, ok := files["manifest.json"]
	if !ok {
		return nil, fmt.Errorf("%w: manifest.json is required at ZIP root", ErrInvalidPackage)
	}
	manifest, err := ParseManifest(manifestData)
	if err != nil {
		return nil, err
	}
	wasmData, ok := files[manifest.Wasm]
	if !ok {
		return nil, fmt.Errorf("%w: wasm %q is missing", ErrInvalidPackage, manifest.Wasm)
	}
	if err := verifyChecksums(files); err != nil {
		return nil, err
	}
	if manifest.UI != nil {
		uiData, exists := files[manifest.UI.Manifest]
		if !exists {
			return nil, fmt.Errorf("%w: UI manifest %q is missing", ErrInvalidPackage, manifest.UI.Manifest)
		}
		uiManifest, parseErr := ParseUIManifest(uiData)
		if parseErr != nil {
			return nil, parseErr
		}
		uiBase := filepath.ToSlash(filepath.Dir(manifest.UI.Manifest))
		if _, exists := files[joinPackagePath(uiBase, uiManifest.Entry)]; !exists {
			return nil, fmt.Errorf("%w: UI entry is missing", ErrInvalidPackage)
		}
		if uiManifest.Style != "" {
			if _, exists := files[joinPackagePath(uiBase, uiManifest.Style)]; !exists {
				return nil, fmt.Errorf("%w: UI style is missing", ErrInvalidPackage)
			}
		}
	}
	if err := validateWasmModule(ctx, wasmData, manifest, limits); err != nil {
		return nil, err
	}
	digest := sha256.Sum256(files["package.sha256"])
	return &ValidatedPackage{
		Manifest: manifest,
		Files:    files,
		Digest:   hex.EncodeToString(digest[:]),
	}, nil
}

func joinPackagePath(base, relative string) string {
	if base == "." || base == "" {
		return relative
	}
	return base + "/" + relative
}

func verifyChecksums(files map[string][]byte) error {
	checksumData, ok := files["package.sha256"]
	if !ok {
		return fmt.Errorf("%w: package.sha256 is required", ErrInvalidPackage)
	}
	expected := make(map[string]string)
	scanner := bufio.NewScanner(bytes.NewReader(checksumData))
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 || len(parts[0]) != sha256.Size*2 {
			return fmt.Errorf("%w: malformed checksum line", ErrInvalidPackage)
		}
		if _, err := hex.DecodeString(parts[0]); err != nil {
			return fmt.Errorf("%w: malformed checksum digest", ErrInvalidPackage)
		}
		if err := validatePackagePath(parts[1]); err != nil {
			return fmt.Errorf("%w: checksum path: %v", ErrInvalidPackage, err)
		}
		if parts[1] == "package.sha256" || parts[1] == "package.sig" {
			return fmt.Errorf("%w: checksum cannot include %q", ErrInvalidPackage, parts[1])
		}
		if _, duplicate := expected[parts[1]]; duplicate {
			return fmt.Errorf("%w: duplicate checksum path %q", ErrInvalidPackage, parts[1])
		}
		expected[parts[1]] = strings.ToLower(parts[0])
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("%w: read checksums: %v", ErrInvalidPackage, err)
	}
	for name, data := range files {
		if name == "package.sha256" || name == "package.sig" {
			continue
		}
		want, exists := expected[name]
		if !exists {
			return fmt.Errorf("%w: checksum missing for %q", ErrInvalidPackage, name)
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != want {
			return fmt.Errorf("%w: checksum mismatch for %q", ErrInvalidPackage, name)
		}
		delete(expected, name)
	}
	if len(expected) != 0 {
		return fmt.Errorf("%w: checksum references missing files", ErrInvalidPackage)
	}
	return nil
}

func validateWasmModule(ctx context.Context, wasmData []byte, manifest *Manifest, limits PackageLimits) error {
	runtime := wazero.NewRuntime(ctx)
	defer runtime.Close(ctx)
	compiled, err := runtime.CompileModule(ctx, wasmData)
	if err != nil {
		return fmt.Errorf("%w: compile wasm: %v", ErrInvalidPackage, err)
	}
	defer compiled.Close(ctx)

	for _, imported := range compiled.ImportedFunctions() {
		moduleName, name, isImport := imported.Import()
		if !isImport {
			return fmt.Errorf("%w: unsupported function import %s.%s", ErrInvalidPackage, moduleName, name)
		}
		switch moduleName {
		case HostModuleName:
			if name != HostCallName {
				return fmt.Errorf("%w: unsupported function import %s.%s", ErrInvalidPackage, moduleName, name)
			}
			if err := validateFunctionSignature(imported, []byte{0x7f, 0x7f, 0x7f, 0x7f}, []byte{0x7f}); err != nil {
				return fmt.Errorf("%w: invalid host_call signature: %v", ErrInvalidPackage, err)
			}
		case WASIModuleName:
			// WASI is instantiated without environment variables, sockets, or
			// preopened files. It only supplies the Go runtime reactor imports.
			if _, allowed := allowedWASIImports[name]; !allowed {
				return fmt.Errorf("%w: unsupported WASI import %s", ErrInvalidPackage, name)
			}
		default:
			return fmt.Errorf("%w: unsupported function import %s.%s", ErrInvalidPackage, moduleName, name)
		}
	}
	if len(compiled.ImportedMemories()) != 0 {
		return fmt.Errorf("%w: imported memory is not allowed", ErrInvalidPackage)
	}
	exports := compiled.ExportedFunctions()
	expected := map[string]struct {
		params  []byte
		results []byte
	}{
		ExportAlloc:   {params: []byte{0x7f}, results: []byte{0x7f}},
		ExportFree:    {params: []byte{0x7f, 0x7f}},
		ExportInit:    {params: []byte{0x7f, 0x7f}, results: []byte{0x7e}},
		ExportOnEvent: {params: []byte{0x7f, 0x7f}, results: []byte{0x7e}},
	}
	for name, signature := range expected {
		definition, exists := exports[name]
		if !exists {
			return fmt.Errorf("%w: required export %q is missing", ErrInvalidPackage, name)
		}
		if err := validateFunctionSignature(definition, signature.params, signature.results); err != nil {
			return fmt.Errorf("%w: invalid export %q signature: %v", ErrInvalidPackage, name, err)
		}
	}
	if containsString(manifest.Permissions, "route.handle") {
		definition, exists := exports[ExportHandleHTTP]
		if !exists {
			return fmt.Errorf("%w: route.handle requires export %q", ErrInvalidPackage, ExportHandleHTTP)
		}
		if err := validateFunctionSignature(definition, []byte{0x7f, 0x7f}, []byte{0x7e}); err != nil {
			return fmt.Errorf("%w: invalid export %q signature: %v", ErrInvalidPackage, ExportHandleHTTP, err)
		}
	}
	if definition, exists := exports[ExportMigrate]; exists {
		if err := validateFunctionSignature(definition, []byte{0x7f, 0x7f}, []byte{0x7e}); err != nil {
			return fmt.Errorf("%w: invalid export %q signature: %v", ErrInvalidPackage, ExportMigrate, err)
		}
	}
	memories := compiled.ExportedMemories()
	memory, exists := memories["memory"]
	if !exists {
		return fmt.Errorf("%w: exported memory is required", ErrInvalidPackage)
	}
	if memory.Min() > limits.MaxInitialMemoryPages {
		return fmt.Errorf("%w: initial memory exceeds %d pages", ErrInvalidPackage, limits.MaxInitialMemoryPages)
	}
	return nil
}

func validateFunctionSignature(definition interface {
	ParamTypes() []api.ValueType
	ResultTypes() []api.ValueType
}, params, results []byte) error {
	actualParams := definition.ParamTypes()
	actualResults := definition.ResultTypes()
	if len(actualParams) != len(params) || len(actualResults) != len(results) {
		return fmt.Errorf("expected %d params and %d results", len(params), len(results))
	}
	for index, expected := range params {
		if byte(actualParams[index]) != expected {
			return fmt.Errorf("parameter %d has unexpected type", index)
		}
	}
	for index, expected := range results {
		if byte(actualResults[index]) != expected {
			return fmt.Errorf("result %d has unexpected type", index)
		}
	}
	return nil
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func sortedFileNames(files map[string][]byte) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func BuildPackage(sourceDir, outputPath string) error {
	files := make(map[string][]byte)
	err := filepath.WalkDir(sourceDir, func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}
		if relative == "." {
			return nil
		}
		name := filepath.ToSlash(relative)
		if entry.IsDir() {
			base := entry.Name()
			if base == "dist" || base == "target" || strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 || !entry.Type().IsRegular() {
			return fmt.Errorf("%w: %q is not a regular file", ErrInvalidPackage, name)
		}
		if name == "package.sha256" || name == "package.sig" ||
			name == "config.json" || name == "state.json" || name == "config.migration.json" ||
			strings.HasSuffix(name, ".panplugin") {
			return nil
		}
		if err := validatePackagePath(name); err != nil {
			return err
		}
		data, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}
		files[name] = data
		return nil
	})
	if err != nil {
		return err
	}
	if _, ok := files["manifest.json"]; !ok {
		return fmt.Errorf("%w: manifest.json is required", ErrInvalidPackage)
	}
	checksums := buildChecksumFile(files)
	files["package.sha256"] = checksums

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(outputPath), ".panplugin-*.tmp")
	if err != nil {
		return err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	zipWriter := zip.NewWriter(tmp)
	names := sortedFileNames(files)
	for _, name := range names {
		header := &zip.FileHeader{Name: name, Method: zip.Deflate}
		header.Modified = deterministicZipTime
		header.SetMode(0o644)
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			tmp.Close()
			return err
		}
		if _, err := writer.Write(files[name]); err != nil {
			tmp.Close()
			return err
		}
	}
	if err := zipWriter.Close(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmpPath, outputPath)
}

func buildChecksumFile(files map[string][]byte) []byte {
	var builder strings.Builder
	names := sortedFileNames(files)
	for _, name := range names {
		sum := sha256.Sum256(files[name])
		fmt.Fprintf(&builder, "%s  %s\n", hex.EncodeToString(sum[:]), name)
	}
	return []byte(builder.String())
}
