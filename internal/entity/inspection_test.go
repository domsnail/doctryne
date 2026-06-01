package entity

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/domsnail/doctryne/pkg/types"
)

func TestInspection_ResolveTarget(t *testing.T) {
	t.Run("resolve target by types", func(t *testing.T) {
		t.Parallel()

		tests := []struct {
			name     string
			scanType types.ScanType
			target   string
			setup    func(t *testing.T) (*Inspection, func())
			wantErr  string
			wantBody string
		}{
			{
				name:     "url",
				scanType: types.ScanType_URL,
				target:   "http://example.test",
				setup: func(t *testing.T) (*Inspection, func()) {
					t.Helper()

					srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						if r.Method != http.MethodGet {
							t.Fatalf("method = %s, want GET", r.Method)
						}
						_, _ = w.Write([]byte("manifest-from-url"))
					}))

					ins := &Inspection{
						ScanType: types.ScanType_URL,
						Target:   bytes.NewBufferString(srv.URL),
						Manifest: &Manifest{},
					}
					return ins, srv.Close
				},
				wantBody: "manifest-from-url",
			},
			{
				name:     "file path",
				scanType: types.ScanType_FilePath,
				target:   "manifest.yaml",
				setup: func(t *testing.T) (*Inspection, func()) {
					t.Helper()

					dir := t.TempDir()
					path := filepath.Join(dir, "manifest.yaml")
					if err := os.WriteFile(path, []byte("manifest-from-file"), 0o600); err != nil {
						t.Fatal(err)
					}

					ins := &Inspection{
						ScanType: types.ScanType_FilePath,
						Target:   bytes.NewBufferString(path),
						Manifest: &Manifest{},
					}
					return ins, func() {}
				},
				wantBody: "manifest-from-file",
			},
			{
				name:     "binary",
				scanType: types.ScanType_Binary,
				target:   "",
				setup: func(t *testing.T) (*Inspection, func()) {
					t.Helper()

					ins := &Inspection{
						ScanType: types.ScanType_Binary,
						Target:   bytes.NewBufferString("unused"),
						Options:  &InspectionOptions{Target: bytes.NewBufferString("manifest-from-binary")},
						Manifest: &Manifest{},
					}
					return ins, func() {}
				},
				wantBody: "manifest-from-binary",
			},
			{
				name:     "empty target",
				scanType: types.ScanType_URL,
				target:   "",
				setup: func(t *testing.T) (*Inspection, func()) {
					t.Helper()
					ins := &Inspection{
						ScanType: types.ScanType_URL,
						Target:   bytes.NewBuffer(nil),
						Manifest: &Manifest{},
					}
					return ins, func() {}
				},
				wantErr: "target is empty",
			},
			{
				name:     "unsupported scan type",
				scanType: types.ScanType("unknown"),
				target:   "anything",
				setup: func(t *testing.T) (*Inspection, func()) {
					t.Helper()
					ins := &Inspection{
						ScanType: types.ScanType("unknown"),
						Target:   bytes.NewBufferString("anything"),
						Manifest: &Manifest{},
					}
					return ins, func() {}
				},
				wantErr: "unspecified scan type",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ins, cleanup := tt.setup(t)
				defer cleanup()

				err := ins.ResolveTarget(context.Background())
				if tt.wantErr != "" {
					if err == nil || err.Error() != tt.wantErr {
						t.Fatalf("err = %v, want %q", err, tt.wantErr)
					}
					return
				}

				if err != nil {
					t.Fatalf("resolve target error = %v", err)
				}

				if string(ins.Manifest.Raw) != tt.wantBody {
					t.Fatalf("body = %q, want %q", string(ins.Manifest.Raw), tt.wantBody)
				}
			})
		}
	})
}
