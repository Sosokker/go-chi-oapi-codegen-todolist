package repository

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Sosokker/todolist-backend/internal/cache"
	"github.com/Sosokker/todolist-backend/internal/domain"
	"github.com/google/uuid"
)

type cachingTagRepository struct {
	next   TagRepository
	cache  cache.Cache
	logger *slog.Logger
}

func NewCachingTagRepository(next TagRepository, cache cache.Cache, logger *slog.Logger) TagRepository {
	return &cachingTagRepository{
		next:   next,
		cache:  cache,
		logger: logger.With("repository", "tag_cache_decorator"),
	}
}

// --- Cache Key Generation ---
func tagCacheKey(userID, tagID uuid.UUID) string {
	return fmt.Sprintf("user:%s:tag:%s", userID, tagID)
}

// --- TagRepository Interface Implementation ---

func (r *cachingTagRepository) Create(ctx context.Context, tag *domain.Tag) (*domain.Tag, error) {
	createdTag, err := r.next.Create(ctx, tag)
	// Invalidate list caches if/when implemented
	return createdTag, err
}

func (r *cachingTagRepository) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Tag, error) {
	cacheKey := tagCacheKey(userID, id)
	if cached, found := r.cache.Get(ctx, cacheKey); found {
		if tag, ok := cached.(*domain.Tag); ok {
			r.logger.DebugContext(ctx, "GetTagByID cache hit", "key", cacheKey)
			return tag, nil
		}
		// Invalid type in cache, treat as miss and delete
		r.logger.WarnContext(ctx, "Invalid type found in tag cache", "key", cacheKey, "type", fmt.Sprintf("%T", cached))
		r.cache.Delete(ctx, cacheKey)
	}

	r.logger.DebugContext(ctx, "GetTagByID cache miss", "key", cacheKey)
	tag, err := r.next.GetByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	if tag != nil {
		r.cache.Set(ctx, cacheKey, tag, 0)
		r.logger.DebugContext(ctx, "Set tag in cache", "key", cacheKey)
	}

	return tag, nil
}

func (r *cachingTagRepository) Update(ctx context.Context, id, userID uuid.UUID, updateData *domain.Tag) (*domain.Tag, error) {
	updatedTag, err := r.next.Update(ctx, id, userID, updateData)
	if err != nil {
		return nil, err
	}
	cacheKey := tagCacheKey(userID, id)
	r.cache.Delete(ctx, cacheKey)
	r.logger.DebugContext(ctx, "Invalidated tag cache after update", "key", cacheKey)
	return updatedTag, nil
}

func (r *cachingTagRepository) Delete(ctx context.Context, id, userID uuid.UUID) error {
	err := r.next.Delete(ctx, id, userID)
	if err != nil {
		return err
	}
	cacheKey := tagCacheKey(userID, id)
	r.cache.Delete(ctx, cacheKey)
	r.logger.DebugContext(ctx, "Invalidated tag cache after delete", "key", cacheKey)
	return nil
}

// --- Pass-through methods ---
func (r *cachingTagRepository) GetByIDs(ctx context.Context, ids []uuid.UUID, userID uuid.UUID) ([]domain.Tag, error) {
	return r.next.GetByIDs(ctx, ids, userID)
}

func (r *cachingTagRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.Tag, error) {
	return r.next.ListByUser(ctx, userID)
}
