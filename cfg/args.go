package cfg

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"

	"github.com/domsnail/doctryne/pkg/types"
)

func NewScanFromArgs(ctx context.Context) (ScanConfig, error) {
	args := flag.Args()

	if len(args) == 0 {
		return ScanConfig{}, errors.New("no target provided")
	}

	if len(args) == 1 {
		return resolveTargetType(args[0])
	}

	var (
		scanType types.ScanType
		targets  = args[1:]
	)

	switch args[0] {
	case "bin", "pipe":
		scanType = types.ScanType_Binary
	case "url", "link", "git":
		scanType = types.ScanType_URL
	case "fs", "file", "path", "files", "filesystem", "filepath":
		scanType = types.ScanType_FilePath
	default:
		return ScanConfig{}, fmt.Errorf("invalid target type: '%s'", args[0])
	}

	slog.DebugContext(ctx, "provided scan target arguments",
		slog.String("scan_type", string(scanType)),
		slog.Any("targets", targets),
	)

	return ScanConfig{
		Targets: targets,
		Type:    scanType,
	}, nil
}

func resolveTargetType(target string) (ScanConfig, error) {
	return ScanConfig{}, errors.New("not implemented")
}
