package article

import (
	"context"

	"github.com/afif0808/testkumparan/app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GetManyFunc func(ctx context.Context, parameters map[string]interface{}) ([]models.Article, error)
type CreateFunc func(ctx context.Context, a models.Article) (models.Article, error)
type GetByIDFunc func(ctx context.Context, id primitive.ObjectID) (*models.Article, error)
type UpdateCacheFunc func(ctx context.Context, art *models.Article, key string) error
