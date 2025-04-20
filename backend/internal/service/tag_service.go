package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/Sosokker/todolist-backend/internal/cache"      // Adjust path
	"github.com/Sosokker/todolist-backend/internal/domain"     // Adjust path
	"github.com/Sosokker/todolist-backend/internal/repository" // Adjust path
	"github.com/google/uuid"
)

type tagService struct {
	tagRepo repository.TagRepository
	cache   cache.Cache // Inject Cache interface
	logger  *slog.Logger
}

// NewTagService creates a new TagService
func NewTagService(repo repository.TagRepository /*, cache cache.Cache */) TagService {
	logger := slog.Default().With("service", "tag")

	var c cache.Cache = nil // Or initialize cache here: cache.NewMemoryCache(...)
	if c != nil {
		logger.Info("TagService initialized with caching enabled")
	} else {
		logger.Info("TagService initialized without caching")
	}

	return &tagService{
		tagRepo: repo,
		cache:   c,
		logger:  logger,
	}
}

func (s *tagService) getCacheKey(tagID uuid.UUID) string {
	// Consider user-specific keys if caching user-scoped data: fmt.Sprintf("user:%s:tag:%s", userID, tagID)
	return "tag:" + tagID.String()
}

func (s *tagService) CreateTag(ctx context.Context, userID uuid.UUID, input CreateTagInput) (*domain.Tag, error) {
	// Use centralized validation
	if err := ValidateCreateTagInput(input); err != nil {
		return nil, err
	}

	tag := &domain.Tag{
		UserID: userID,
		Name:   input.Name,
		Color:  input.Color,
		Icon:   input.Icon,
	}

	createdTag, err := s.tagRepo.Create(ctx, tag)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			s.logger.WarnContext(ctx, "Tag creation conflict", "userId", userID, "tagName", input.Name)
			return nil, err
		}
		s.logger.ErrorContext(ctx, "Failed to create tag in repo", "error", err, "userId", userID)
		return nil, domain.ErrInternalServer
	}

	s.logger.InfoContext(ctx, "Tag created successfully", "tagId", createdTag.ID, "userId", userID)
	return createdTag, nil
}

func (s *tagService) GetTagByID(ctx context.Context, tagID, userID uuid.UUID) (*domain.Tag, error) {
	cacheKey := s.getCacheKey(tagID)

	if s.cache != nil {
		if cachedTag, found := s.cache.Get(ctx, cacheKey); found {
			if tag, ok := cachedTag.(*domain.Tag); ok {
				// IMPORTANT: Verify ownership even on cache hit
				if tag.UserID != userID {
					s.logger.WarnContext(ctx, "Cache hit for tag owned by different user", "tagId", tagID, "ownerId", tag.UserID, "requesterId", userID)
					return nil, domain.ErrNotFound
				}
				s.logger.DebugContext(ctx, "GetTagByID cache hit", "tagId", tagID, "userId", userID)
				return tag, nil
			} else {
				s.logger.WarnContext(ctx, "Invalid type found in tag cache", "key", cacheKey)
				s.cache.Delete(ctx, cacheKey)
			}
		} else {
			s.logger.DebugContext(ctx, "GetTagByID cache miss", "tagId", tagID, "userId", userID)
		}
	}

	tag, err := s.tagRepo.GetByID(ctx, tagID, userID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			s.logger.WarnContext(ctx, "Tag not found by ID in repo", "tagId", tagID, "userId", userID)
		} else {
			s.logger.ErrorContext(ctx, "Failed to get tag by ID from repo", "error", err, "tagId", tagID, "userId", userID)
			err = domain.ErrInternalServer
		}
		return nil, err
	}

	if s.cache != nil {
		s.cache.Set(ctx, cacheKey, tag, 0)
		s.logger.DebugContext(ctx, "Set tag in cache", "tagId", tagID)
	}

	return tag, nil
}

func (s *tagService) ListUserTags(ctx context.Context, userID uuid.UUID) ([]domain.Tag, error) {
	s.logger.DebugContext(ctx, "Listing user tags", "userId", userID)

	tags, err := s.tagRepo.ListByUser(ctx, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to list user tags from repo", "error", err, "userId", userID)
		return nil, domain.ErrInternalServer
	}

	s.logger.DebugContext(ctx, "Found user tags", "userId", userID, "count", len(tags))
	return tags, nil
}

func (s *tagService) UpdateTag(ctx context.Context, tagID, userID uuid.UUID, input UpdateTagInput) (*domain.Tag, error) {
	// Get existing tag first to ensure it exists and belongs to user
	existingTag, err := s.GetTagByID(ctx, tagID, userID)
	if err != nil {
		return nil, err
	}

	if err := ValidateUpdateTagInput(input); err != nil {
		return nil, err
	}

	updateData := &domain.Tag{
		Name:  existingTag.Name,
		Color: existingTag.Color,
		Icon:  existingTag.Icon,
	}
	needsUpdate := false

	if input.Name != nil {
		updateData.Name = *input.Name
		needsUpdate = true
	}
	if input.Color != nil {
		updateData.Color = input.Color
		needsUpdate = true
	}
	if input.Icon != nil {
		updateData.Icon = input.Icon
		needsUpdate = true
	}

	if !needsUpdate {
		s.logger.InfoContext(ctx, "No fields provided for tag update", "tagId", tagID, "userId", userID)
		return existingTag, nil
	}

	updatedTag, err := s.tagRepo.Update(ctx, tagID, userID, updateData)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			s.logger.WarnContext(ctx, "Tag update conflict", "error", err, "tagId", tagID, "userId", userID, "conflictingName", updateData.Name)
			return nil, err
		}
		s.logger.ErrorContext(ctx, "Failed to update tag in repo", "error", err, "tagId", tagID, "userId", userID)
		return nil, domain.ErrInternalServer
	}

	if s.cache != nil {
		cacheKey := s.getCacheKey(tagID)
		s.cache.Delete(ctx, cacheKey)
		s.logger.DebugContext(ctx, "Invalidated tag cache after update", "tagId", tagID)
	}

	s.logger.InfoContext(ctx, "Tag updated successfully", "tagId", tagID, "userId", userID)
	return updatedTag, nil
}

func (s *tagService) DeleteTag(ctx context.Context, tagID, userID uuid.UUID) error {
	_, err := s.GetTagByID(ctx, tagID, userID)
	if err != nil {
		return err
	}

	err = s.tagRepo.Delete(ctx, tagID, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to delete tag from repo", "error", err, "tagId", tagID, "userId", userID)
		return domain.ErrInternalServer
	}

	if s.cache != nil {
		cacheKey := s.getCacheKey(tagID)
		s.cache.Delete(ctx, cacheKey)
		s.logger.DebugContext(ctx, "Invalidated tag cache after delete", "tagId", tagID)
	}

	s.logger.InfoContext(ctx, "Tag deleted successfully", "tagId", tagID, "userId", userID)
	return nil
}

// ValidateUserTags checks if all provided tag IDs exist and belong to the user.
func (s *tagService) ValidateUserTags(ctx context.Context, userID uuid.UUID, tagIDs []uuid.UUID) error {
	if len(tagIDs) == 0 {
		return nil
	}

	// GetByIDs repo method already filters by userID
	foundTags, err := s.tagRepo.GetByIDs(ctx, tagIDs, userID)
	if err != nil {
		s.logger.ErrorContext(ctx, "Failed to get tags by IDs during validation", "error", err, "userId", userID)
		return domain.ErrInternalServer
	}

	if len(foundTags) != len(tagIDs) {
		foundMap := make(map[uuid.UUID]bool)
		for _, t := range foundTags {
			foundMap[t.ID] = true
		}
		missing := []string{}
		for _, reqID := range tagIDs {
			if !foundMap[reqID] {
				missing = append(missing, reqID.String())
			}
		}
		errMsg := "invalid or forbidden tag IDs: " + strings.Join(missing, ", ")
		s.logger.WarnContext(ctx, "Tag validation failed", "userId", userID, "missingTags", missing)
		return fmt.Errorf("%s: %w", errMsg, domain.ErrBadRequest)
	}

	s.logger.DebugContext(ctx, "User tags validated successfully", "userId", userID, "tagCount", len(tagIDs))
	return nil
}
