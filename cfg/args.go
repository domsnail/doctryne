package cfg

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"

	"github.com/domsnail/doctryne/pkg/types"
)

func NewScanFromArgs(ctx context.Context) (*Scan, error) {
	args := flag.Args()

	if len(args) == 0 {
		return nil, errors.New("no target provided")
	}

	if len(args) == 1 {
		return resolveTargetType(args[0])
	}

	var (
		scanType types.ScanType
		targets  = args[1:]
	)

	switch args[0] {
	case "cdx", "cyclonedx", "sbom":
		scanType = types.ScanType_CycloneDX
	case "fs", "file", "files", "filesystem":
		scanType = types.ScanType_Files
		return nil, fmt.Errorf("target type not implemented")
	default:
		return nil, fmt.Errorf("invalid target type: '%s'", args[0])
	}

	slog.DebugContext(ctx, "provided scan target arguments",
		slog.String("scan_type", string(scanType)),
		slog.Any("targets", targets),
	)

	return &Scan{
		Targets: targets,
		Type:    scanType,
	}, nil
}

func resolveTargetType(target string) (*Scan, error) {
	return nil, errors.New("not implemented")
}
