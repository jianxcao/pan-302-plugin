package pluginpkg

import (
	"encoding/json"
	"fmt"
	"path"
	"regexp"
	"strings"
)

var (
	pluginNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)
	versionPattern    = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:[-+][0-9A-Za-z.-]+)?$`)
)

type Manifest struct {
	SchemaVersion int      `json:"schemaVersion"`
	Name          string   `json:"name"`
	Version       string   `json:"version"`
	DisplayName   string   `json:"displayName"`
	Description   string   `json:"description,omitempty"`
	Author        string   `json:"author,omitempty"`
	ABI           string   `json:"abi"`
	Wasm          string   `json:"wasm"`
	ConfigVersion int      `json:"configVersion,omitempty"`
	Events        []string `json:"events,omitempty"`
	Permissions   []string `json:"permissions,omitempty"`
	UI            *UIRef   `json:"ui,omitempty"`
}

type UIRef struct {
	Manifest string `json:"manifest"`
	Mode     string `json:"mode"`
}

type UIManifest struct {
	SchemaVersion int    `json:"schemaVersion"`
	Name          string `json:"name"`
	Version       string `json:"version"`
	Protocol      string `json:"protocol"`
	Mode          string `json:"mode"`
	Entry         string `json:"entry"`
	Style         string `json:"style,omitempty"`
	ExportName    string `json:"exportName,omitempty"`
}

func ParseManifest(data []byte) (*Manifest, error) {
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("%w: JSON: %v", ErrInvalidManifest, err)
	}
	if err := manifest.Validate(); err != nil {
		return nil, err
	}
	return &manifest, nil
}

func (m *Manifest) Validate() error {
	if m.SchemaVersion != 1 {
		return fmt.Errorf("%w: unsupported schemaVersion %d", ErrInvalidManifest, m.SchemaVersion)
	}
	if !pluginNamePattern.MatchString(m.Name) {
		return fmt.Errorf("%w: invalid name %q", ErrInvalidManifest, m.Name)
	}
	if !versionPattern.MatchString(m.Version) {
		return fmt.Errorf("%w: invalid version %q", ErrInvalidManifest, m.Version)
	}
	if strings.TrimSpace(m.DisplayName) == "" {
		return fmt.Errorf("%w: displayName is required", ErrInvalidManifest)
	}
	if m.ABI != ABIVersion {
		return fmt.Errorf("%w: unsupported ABI %q", ErrInvalidManifest, m.ABI)
	}
	if err := validatePackagePath(m.Wasm); err != nil {
		return fmt.Errorf("%w: wasm: %v", ErrInvalidManifest, err)
	}
	if !strings.HasSuffix(strings.ToLower(m.Wasm), ".wasm") {
		return fmt.Errorf("%w: wasm must use .wasm extension", ErrInvalidManifest)
	}
	if m.ConfigVersion < 0 {
		return fmt.Errorf("%w: configVersion must not be negative", ErrInvalidManifest)
	}
	for _, event := range m.Events {
		switch event {
		case "strm.created",
			"strm.overwritten",
			"strm.deleted",
			"strm.renamed",
			"strm.moved",
			"strm.copied",
			"media.item.added",
			"media.item.deleted":
		default:
			return fmt.Errorf("%w: unsupported event %q", ErrInvalidManifest, event)
		}
	}
	for _, permission := range m.Permissions {
		if _, ok := CoreHostOperations[permission]; !ok {
			return fmt.Errorf("%w: unsupported permission %q", ErrInvalidManifest, permission)
		}
	}
	if m.UI != nil {
		if m.UI.Mode != "trusted-esm" {
			return fmt.Errorf("%w: unsupported UI mode %q", ErrInvalidManifest, m.UI.Mode)
		}
		if err := validatePackagePath(m.UI.Manifest); err != nil {
			return fmt.Errorf("%w: UI manifest: %v", ErrInvalidManifest, err)
		}
	}
	return nil
}

func (m *Manifest) EffectiveConfigVersion() int {
	if m.ConfigVersion == 0 {
		return 1
	}
	return m.ConfigVersion
}

func ParseUIManifest(data []byte) (*UIManifest, error) {
	var manifest UIManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("%w: UI JSON: %v", ErrInvalidManifest, err)
	}
	if manifest.SchemaVersion != 1 ||
		manifest.Protocol != UIProtocol ||
		manifest.Mode != "trusted-esm" ||
		!pluginNamePattern.MatchString(manifest.Name) ||
		!versionPattern.MatchString(manifest.Version) {
		return nil, fmt.Errorf("%w: invalid UI manifest metadata", ErrInvalidManifest)
	}
	if err := validatePackagePath(manifest.Entry); err != nil {
		return nil, fmt.Errorf("%w: UI entry: %v", ErrInvalidManifest, err)
	}
	if manifest.Style != "" {
		if err := validatePackagePath(manifest.Style); err != nil {
			return nil, fmt.Errorf("%w: UI style: %v", ErrInvalidManifest, err)
		}
	}
	return &manifest, nil
}

func validatePackagePath(value string) error {
	if value == "" {
		return fmt.Errorf("path is required")
	}
	if strings.Contains(value, `\`) || strings.HasPrefix(value, "/") {
		return fmt.Errorf("path must be a forward-slash relative path")
	}
	clean := path.Clean(value)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || clean != value {
		return fmt.Errorf("path is not canonical")
	}
	return nil
}
