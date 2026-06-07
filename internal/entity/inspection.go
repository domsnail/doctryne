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
	"github.com/domsnail/doctryne/pkg/utils"
	"github.com/google/uuid"
)

// Inspection is a resulting entity for a scan of a target (one inspection = multiple manifests) from API or CLI.
// If Inspection runs in dir mode multiple manifests can be found
type Inspection struct {
	UUID uuid.UUID

	// target can be a binary file, url, package or module name
	Target   io.Reader
	Lockfile io.Reader
	ScanType types.ScanType

	Manifest *Manifest // todo: add multiple manifests support

	Packages     []*Package
	Repositories []*Repository
	Developers   []*Developer

	Options *InspectionOptions
}

type InspectionOptions struct {
	ScanType types.ScanType
	Target   io.Reader
	Lockfile io.Reader

	Mode types.InspectionMode

	// LoadUserProfiles is set to true, all users profiles (contributors) will be additionally
	// loaded (their profile, activity, repositories)
	LoadUserProfiles bool
}

func NewInspection(opts *InspectionOptions) *Inspection {
	ins := Inspection{
		UUID:     uuid.Must(uuid.NewV7()),
		Target:   opts.Target,
		Lockfile: opts.Lockfile,
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

	ins.Manifest.WithAuthor(utils.GetClientDataFromIncomingMetadata(ctx))

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

		err = ins.Manifest.WithFilename(buf.String()).SetFileContent(resp.Body)
		if err != nil {
			return fmt.Errorf("failed to read manifest body contents: %w", err)
		}

	case types.ScanType_FilePath:
		file, err := os.ReadFile(buf.String())
		if err != nil {
			return fmt.Errorf("failed to read from file: %w", err)
		}

		ins.Manifest.Metadata.Filename = filepath.Base(buf.String())
		err = ins.Manifest.WithFilename(buf.String()).SetFileContent(bytes.NewReader(file))
		if err != nil {
			return fmt.Errorf("failed to read manifest body contents: %w", err)
		}
	case types.ScanType_DirPath:
		// todo: search for files
		return errors.New("not implemented")
	case types.ScanType_Binary:
		err = ins.Manifest.SetFileContent(ins.Options.Target)
		if err != nil {
			return fmt.Errorf("failed to read manifest body contents: %w", err)
		}

	default:
		return errors.New("unspecified scan type")
	}

	if ins.Lockfile != nil {
		var lbuf bytes.Buffer
		n, err = lbuf.ReadFrom(ins.Lockfile)
		if err != nil {
			return fmt.Errorf("failed to read lockfile: %w", err)
		} else if n == 0 {
			return errors.New("lockfile is empty")
		}

		switch ins.ScanType {
		case types.ScanType_URL:
			u, err := url.Parse(lbuf.String())
			if err != nil {
				return fmt.Errorf("failed to parse lockfile url: %w", err)
			}

			request, err := http.NewRequest(http.MethodGet, u.String(), nil)
			if err != nil {
				return err
			}

			slog.DebugContext(ctx, "querying lockfile url...",
				slog.String("scan_type", string(ins.ScanType)),
				slog.String("url", u.Redacted()),
			)

			resp, err := http.DefaultClient.Do(request.WithContext(ctx))
			if err != nil {
				return fmt.Errorf("failed to perform http request: %w", err)
			}

			err = ins.Manifest.SetLockfileContent(resp.Body)
			if err != nil {
				return fmt.Errorf("failed to read lockfile body contents: %w", err)
			}
		case types.ScanType_DirPath:
			// todo: search for files
			return errors.New("not implemented")
		case types.ScanType_FilePath:
			file, err := os.ReadFile(lbuf.String())
			if err != nil {
				return fmt.Errorf("failed to read from lockfile: %w", err)
			}

			err = ins.Manifest.SetLockfileContent(bytes.NewReader(file))
			if err != nil {
				return fmt.Errorf("failed to read lockfile body contents: %w", err)
			}
		}
	}

	return nil
}
