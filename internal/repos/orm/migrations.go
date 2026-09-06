package orm

import (
	"embed"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/domsnail/doctryne/internal/models"
	"gorm.io/gorm"
)

//go:embed _embed
var embeddedScriptFiles embed.FS

func AutoMigrate(db *gorm.DB) error {
	err := CreateExtensions(db)
	if err != nil {
		return err
	}

	err = MigrateTypes(db)
	if err != nil {
		return err
	}

	err = db.AutoMigrate(
		&models.InspectionModel{},
		&models.DeveloperModel{},
	)

	if err != nil {
		return err
	}

	err = db.AutoMigrate(
		&models.VulnerabilityModel{},
		&models.VulnerabilitySourceModel{},
		&models.VulnerabilityScoreModel{},
		&models.VulnerabilityReferenceModel{},
		&models.VulnerabilityDescriptionModel{},
		&models.VulnerabilityCanonicalModel{},
	)

	if err != nil {
		return err
	}

	err = RunPrefills(db)
	if err != nil {
		return err
	}

	return err
}

func CreateExtensions(db *gorm.DB) error {
	scripts, err := prepareScripts("_embed/extensions")
	if err != nil {
		return err
	}

	slog.Info("updating extensions...",
		slog.Int("total_files", len(scripts)),
	)

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, s := range scripts {
			err = tx.Exec(s).Error
			if err != nil {
				return fmt.Errorf("failed to execute embedded file script: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("failed to update extensions",
			slog.String("error", err.Error()),
		)

		return err
	}

	slog.Info("extensions updated successfully")
	return nil
}

func MigrateTypes(db *gorm.DB) error {
	scripts, err := prepareScripts("_embed/types")
	if err != nil {
		return err
	}

	slog.Info("migrating data types...",
		slog.Int("total_files", len(scripts)),
	)

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, s := range scripts {
			err = tx.Exec(s).Error
			if err != nil {
				return fmt.Errorf("failed to execute embedded file script: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("failed to migrate data types",
			slog.String("error", err.Error()),
		)

		return err
	}

	slog.Info("data types migrated successfully")
	return nil
}

func RunPrefills(db *gorm.DB) error {
	scripts, err := prepareScripts("_embed/data")
	if err != nil {
		return err
	}

	slog.Info("running data prefills...",
		slog.Int("total_files", len(scripts)),
	)

	err = db.Transaction(func(tx *gorm.DB) error {
		for _, s := range scripts {
			err = tx.Exec(s).Error
			if err != nil {
				return fmt.Errorf("failed to execute embedded file script: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		slog.Error("failed to prefill data",
			slog.String("error", err.Error()),
		)

		return err
	}

	slog.Info("data prefilled successfully")
	return nil
}

func prepareScripts(path string) (scripts []string, err error) {
	files, err := embeddedScriptFiles.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded types dir: %w", err)
	} else if len(files) == 0 {
		return nil, fmt.Errorf("embedded types dir contains no files")
	}

	scripts = make([]string, len(files))

	for i, file := range files {
		info, err := file.Info()
		if err != nil {
			return nil, fmt.Errorf("failed to get file info: %w", err)
		}

		if info.IsDir() {
			return nil, fmt.Errorf("embedded types dir contains a directory")
		}

		if filepath.Ext(info.Name()) != ".sql" {
			return nil, fmt.Errorf("embedded types dir contains an unsupported file type")
		}

		script, err := embeddedScriptFiles.ReadFile(filepath.ToSlash(filepath.Join(path, file.Name())))
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded file: %w", err)
		}

		scripts[i] = string(script)
	}

	return
}
