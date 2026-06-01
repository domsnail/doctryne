package entity

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/domsnail/doctryne/pkg/types"
	"github.com/google/uuid"
)

// Inspection is a resulting entity for a single scan of a target (one manifest = one inspection) from API or CLI
type Inspection struct {
	UUID uuid.UUID

	// target can be a binary file, url, package or module name
	Target   io.Reader
	ScanType types.ScanType

	Manifest *Manifest

	Packages     []*Package
	Repositories []*Repository
	Developers   []*Developer

	Options *InspectionOptions
}

type InspectionOptions struct {
	ScanType types.ScanType
	Target   io.Reader

	Mode types.InspectionMode

	// LoadUserProfiles is set to true, all users profiles (contributors) will be additionally
	// loaded (their profile, activity, repositories)
	LoadUserProfiles bool
}

func NewInspection(opts *InspectionOptions) *Inspection {
	ins := Inspection{
		UUID:     uuid.Must(uuid.NewV7()),
		Target:   opts.Target,
		ScanType: opts.ScanType,
		Options:  opts,
		Manifest: NewManifest(),
	}

	return &ins
}

func (ins *Inspection) ResolveTarget(ctx context.Context) error {
	var buf bytes.Buffer
	n, err := buf.ReadFrom(ins.Target)
	if err != nil {
		return fmt.Errorf("failed to read target: %w", err)
	} else if n == 0 {
		return errors.New("target is empty")
	}

	slog.DebugContext(ctx, "resolving target...",
		slog.String("scan_type", string(ins.ScanType)),
		slog.Int64("bytes_read", n),
	)

	switch ins.ScanType {
	case types.ScanType_URL:
		u, err := url.Parse(buf.String())
		if err != nil {
			return fmt.Errorf("failed to parse target url: %w", err)
		}

		request, err := http.NewRequest(http.MethodGet, u.String(), nil)
		if err != nil {
			return err
		}

		slog.DebugContext(ctx, "querying target url...",
			slog.String("scan_type", string(ins.ScanType)),
			slog.String("url", u.Redacted()),
		)

		resp, err := http.DefaultClient.Do(request.WithContext(ctx))
		if err != nil {
			return fmt.Errorf("failed to perform http request: %w", err)
		}

		err = ins.Manifest.SetFileContent(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read manifest body contents: %w", err)
		}

	case types.ScanType_FilePath:
		file, err := os.ReadFile(buf.String())
		if err != nil {
			return fmt.Errorf("failed to read from file: %w", err)
		}

		ins.Manifest.Metadata.Filename = filepath.Base(buf.String())
		err = ins.Manifest.SetFileContent(bytes.NewReader(file))
		if err != nil {
			return fmt.Errorf("failed to read manifest body contents: %w", err)
		}

	case types.ScanType_Binary:
		err = ins.Manifest.SetFileContent(ins.Options.Target)
		if err != nil {
			return fmt.Errorf("failed to read manifest body contents: %w", err)
		}

	default:
		return errors.New("unspecified scan type")
	}

	return nil
}
