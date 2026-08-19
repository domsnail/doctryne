package repos

import (
	"context"
	"errors"
	"strings"

	"github.com/domsnail/doctryne/internal/entity"
	"github.com/domsnail/doctryne/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DevelopersRepoImpl struct {
	orm *gorm.DB
}

func NewDevelopersRepoImpl(orm *gorm.DB) *DevelopersRepoImpl {
	return &DevelopersRepoImpl{orm: orm}
}

func (repo *DevelopersRepoImpl) CreateDeveloper(ctx context.Context, developer *entity.Developer) error {
	if developer == nil {
		return errors.New("developer is nil")
	}

	model := models.NewDeveloperModel(developer)
	err := repo.orm.WithContext(ctx).Create(model).Error
	if err != nil {
		return err
	}

	return nil
}

func (repo *DevelopersRepoImpl) CreateDevelopers(ctx context.Context, developers []*entity.Developer) error {
	if developers == nil || len(developers) == 0 {
		return errors.New("developer slice is nil")
	}

	var model = make([]*models.DeveloperModel, len(developers))
	for i, d := range developers {
		d.UUID = uuid.Must(uuid.NewV7())
		model[i] = models.NewDeveloperModel(d)
	}

	query := repo.orm.WithContext(ctx).CreateInBatches(model, DEFAULT_BATCH_SIZE)
	if query.Error != nil {
		return query.Error
	}

	for i := range developers {
		developers[i].CreatedAt = model[i].CreatedAt
		developers[i].UpdatedAt = model[i].UpdatedAt
	}

	return nil
}

func (repo *DevelopersRepoImpl) UpsertDeveloper(ctx context.Context, developer *entity.Developer) (error, int64) {
	if developer == nil {
		return errors.New("developer is nil"), 0
	}

	model := models.NewDeveloperModel(developer)
	query := repo.orm.WithContext(ctx).FirstOrCreate(model) // todo: use assign
	if query.Error != nil {
		return query.Error, query.RowsAffected
	}

	return nil, query.RowsAffected
}

func (repo *DevelopersRepoImpl) UpsertDevelopers(ctx context.Context, developers []*entity.Developer) (error, int64) {
	if developers == nil {
		return errors.New("developers slice is nil"), 0
	}

	var model = make([]*models.DeveloperModel, len(developers))
	for i, d := range developers {
		model[i] = models.NewDeveloperModel(d)
	}

	query := repo.orm.
		WithContext(ctx).
		Save(model)

	if query.Error != nil {
		return query.Error, query.RowsAffected
	}

	return nil, query.RowsAffected
}

func (repo *DevelopersRepoImpl) SelectDeveloperByUUID(ctx context.Context, uid uuid.UUID) (*entity.Developer, error) {
	if uid.String() == "" {
		return nil, errors.New("uuid is empty")
	}

	var model models.DeveloperModel
	err := repo.orm.WithContext(ctx).Where("uuid = ?", uid).First(&model).Error
	if err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (repo *DevelopersRepoImpl) SelectDevelopersByQueryFilter(ctx context.Context, filter entity.DevelopersQueryFilter) ([]*entity.Developer, error) {
	//TODO implement me
	panic("implement me")
}

// FindOrCreateDevelopers queries already existing developers and creates new
// 1. try to find existing usernames
// 2. try to find existing full names
// 3. todo: try to find existing emails (if none found)
// 4. add new data to existing developers
// 5. update all developers (updated and new)
func (repo *DevelopersRepoImpl) FindOrCreateDevelopers(ctx context.Context, developers []*entity.Developer) (error, int64) {
	if len(developers) == 0 {
		return nil, 0
	}

	var (
		usernames   []string
		usernameMap = make(map[string]*entity.Developer)

		fullnames   []string
		fullnameMap = make(map[string]*entity.Developer)
		//emails    []string

		githubIDs []int64
		githubMap = make(map[int64]*entity.Developer)

		model []*models.DeveloperModel
	)

	for _, developer := range developers {
		if developer.Username != "" {
			usernames = append(usernames, strings.ToLower(developer.Username))
			usernameMap[developer.Username] = developer
		}

		if developer.Name != "" {
			fullnames = append(fullnames, developer.Name)
			fullnameMap[developer.Name] = developer
		}

		//if len(developer.Emails) > 0 {
		//	emails = append(emails, developer.Emails...)
		//}

		if developer.GithubID != nil && *developer.GithubID != 0 {
			githubIDs = append(githubIDs, *developer.GithubID)
			githubMap[*developer.GithubID] = developer
		}
	}

	query := repo.orm.WithContext(ctx)

	if len(usernames) > 0 {
		query = query.Or("username IN ?", usernames)
	}

	if len(fullnames) > 0 {
		query = query.Or("full_name IN ?", fullnames)
	}

	if len(githubIDs) > 0 {
		query = query.Or("github_id IN ?", githubIDs)
	}

	//if len(emails) > 0 {
	//	query = query.Where("full_name IN ?", fullnames)
	//}

	query = query.Order("updated_at DESC")

	result := query.FindInBatches(&model, DEFAULT_BATCH_SIZE, func(tx *gorm.DB, batch int) error {
		return nil
	})

	for _, m := range model {
		if m.GithubID != nil && *m.GithubID != 0 {
			found, ok := githubMap[*m.GithubID]
			if ok {
				// todo: maybe load previous information
				found.UUID = m.UUID

				found.CreatedAt = m.CreatedAt
				found.UpdatedAt = m.UpdatedAt
				continue
			}
		}

		if m.Username.Valid {
			found, ok := usernameMap[m.Username.String]
			if ok {
				found.UUID = m.UUID

				found.CreatedAt = m.CreatedAt
				found.UpdatedAt = m.UpdatedAt
				continue
			}
		}

		if m.FullName.Valid {
			found, ok := fullnameMap[m.FullName.String]
			if ok {
				found.UUID = m.UUID

				found.CreatedAt = m.CreatedAt
				found.UpdatedAt = m.UpdatedAt
				continue
			}
		}
	}

	var developersToCreate []*entity.Developer
	for _, d := range developers {
		if d.UUID != uuid.Nil {
			continue
		}

		developersToCreate = append(developersToCreate, d)
	}

	if len(developersToCreate) > 0 {
		err := repo.CreateDevelopers(ctx, developersToCreate)
		if err != nil {
			return err, 0
		}
	}

	return nil, result.RowsAffected + int64(len(developersToCreate))
}
