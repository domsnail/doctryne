package cyclonedx_parser

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"slices"
	"strings"
	"time"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/pkg/types"
	"github.com/package-url/packageurl-go"
)

type Parser struct {
	ctx context.Context

	file io.Reader
}

func (p *Parser) WithContext(ctx context.Context) *Parser {
	p.ctx = ctx
	return p
}

func (p *Parser) WithFile(file io.Reader) *Parser {
	p.file = file
	return p
}

func (p *Parser) ParseManifest() (*entity.Package, error) {
	if p.ctx.Err() != nil {
		return nil, p.ctx.Err()
	} else if p.file == nil {
		return nil, errors.New("manifest file is nil")
	}

	slog.DebugContext(p.ctx, "decoding cyclonedx sbom file...")

	var bom cdx.BOM
	decoder := cdx.NewBOMDecoder(p.file, cdx.BOMFileFormatJSON)
	err := decoder.Decode(&bom)
	if err != nil {
		return nil, fmt.Errorf("failed to decode cyclonedx sbom file: %w", err)
	}

	var (
		createdAt time.Time

		vulnerabilities int
		components      int
	)

	if bom.Metadata == nil {
		slog.WarnContext(p.ctx, "cyclonedx sbom file has no metadata",
			slog.String("bom_serial_number", bom.SerialNumber),
		)
	} else {
		createdAt, _ = time.Parse(time.RFC3339, bom.Metadata.Timestamp)
	}

	if bom.Components == nil {
		slog.WarnContext(p.ctx, "cyclonedx sbom file has no components",
			slog.String("bom_serial_number", bom.SerialNumber),
		)
	} else {
		components = len(*bom.Components)
	}

	if bom.Vulnerabilities != nil {
		vulnerabilities = len(*bom.Vulnerabilities)
	}

	slog.InfoContext(p.ctx, "successfully decoded cyclonedx sbom file",
		slog.String("serial_number", bom.SerialNumber),
		slog.Int("components", components),
		slog.Int("vulnerabilities", vulnerabilities),
		slog.Time("created_at", createdAt),
	)

	if components == 0 {
		return &entity.Package{}, nil
	}

	rootPkg := rootComponent(&bom)
	rootPkg.Dependencies = make([]*entity.Package, len(*bom.Components))

	for i, component := range *bom.Components {
		rootPkg.Dependencies[i] = convertComponent(&component)
	}

	return rootPkg, nil
}

func rootComponent(bom *cdx.BOM) *entity.Package {
	if bom.Metadata == nil || bom.Metadata.Component == nil {
		slog.Warn("unable to find root component, creating placeholder package...")
		return &entity.Package{
			Name: "placeholder-root-package",
			Labels: []types.Label{
				types.Label_Root,
				types.Label_NonExistant,
			},
			IsDev:      false,
			IsOptional: false,
		}
	}

	c := convertComponent(bom.Metadata.Component)
	c.Labels = append(c.Labels, types.Label_Root)
	return c
}

func convertComponent(c *cdx.Component) *entity.Package {
	var err error

	pkg := entity.Package{
		Name:       strings.Join([]string{c.Group, c.Name}, "/"),
		Version:    c.Version,
		Resolved:   nil, // todo: maybe check cdxgen evince mode
		Registry:   "",
		Labels:     []types.Label{},
		IsOptional: c.Scope == cdx.ScopeOptional,
		CPE:        c.CPE,
	}

	dev := isDevComponent(c.Properties)
	if dev {
		pkg.Labels = append(pkg.Labels, types.Label_Dev)
		pkg.IsDev = true
	}

	pkg.PackageURL, err = packageurl.FromString(c.PackageURL)
	if err != nil {
		slog.Warn("failed to parse component purl",
			slog.String("package_name", c.Name),
			slog.String("package_url", c.PackageURL),
			slog.String("error", err.Error()),
		)
	} else {
		pkg.Ecosystem = types.Ecosystem(pkg.PackageURL.Type)
		// todo: ecosystem to language
	}

	if c.Type != "" {
		v, ok := componentTypeLabel[c.Type]
		if ok {
			pkg.Labels = append(pkg.Labels, v)
		} else {
			slog.Warn("unknown sbom component type encountered",
				slog.String("package_name", c.Name),
				slog.String("component_type", string(c.Type)),
			)
		}
	}

	if c.Hashes != nil && len(*c.Hashes) > 0 {
		h := *c.Hashes
		pkg.Integrity = fmt.Sprintf("%s:%s", strings.ToLower(string(h[1].Algorithm)), h[0].Value)
	}

	return &pkg
}

var componentTypeLabel = map[cdx.ComponentType]types.Label{
	cdx.ComponentTypeApplication:          types.Label_ComponentType_Application,
	cdx.ComponentTypeContainer:            types.Label_ComponentType_Container,
	cdx.ComponentTypeCryptographicAsset:   types.Label_ComponentType_CryptographicAsset,
	cdx.ComponentTypeData:                 types.Label_ComponentType_Data,
	cdx.ComponentTypeDevice:               types.Label_ComponentType_Device,
	cdx.ComponentTypeDeviceDriver:         types.Label_ComponentType_DeviceDriver,
	cdx.ComponentTypeFile:                 types.Label_ComponentType_File,
	cdx.ComponentTypeFirmware:             types.Label_ComponentType_Firmware,
	cdx.ComponentTypeFramework:            types.Label_ComponentType_Framework,
	cdx.ComponentTypeLibrary:              types.Label_ComponentType_Library,
	cdx.ComponentTypeMachineLearningModel: types.Label_ComponentType_MachineLearningModel,
	cdx.ComponentTypeOS:                   types.Label_ComponentType_OS,
	cdx.ComponentTypePlatform:             types.Label_ComponentType_Platform,
}

var devProperties = map[string]bool{
	"cdx:npm:package:development": true,
}

func isDevComponent(props *[]cdx.Property) bool {
	if props == nil || len(*props) == 0 {
		return false
	}

	return slices.ContainsFunc(*props, func(p cdx.Property) bool {
		return devProperties[p.Value] && p.Value == "true"
	})
}
