package store

import (
	"context"
	"errors"

	"github.com/usememos/memos/internal/base"
)

// TravelStoryPayload represents the JSON payload for a travel story.
type TravelStoryPayload struct {
	Tags     []string              `json:"tags,omitempty"`
	Sections []*TravelStorySection `json:"sections,omitempty"`
	Theme    *TravelStoryTheme     `json:"theme,omitempty"`
}

// TravelStorySection represents a section in a travel story.
type TravelStorySection struct {
	Title         string              `json:"title,omitempty"`
	Content       string              `json:"content,omitempty"`
	Date          int64               `json:"date,omitempty"`
	Location      string              `json:"location,omitempty"`
	Images        []*TravelStoryImage `json:"images,omitempty"`
	SourceMemoUID string              `json:"sourceMemoUid,omitempty"`
}

// TravelStoryImage represents an image in a travel story section.
type TravelStoryImage struct {
	URL          string `json:"url,omitempty"`
	Caption      string `json:"caption,omitempty"`
	AltText      string `json:"altText,omitempty"`
	QualityScore int32  `json:"qualityScore,omitempty"`
	Selected     bool   `json:"selected,omitempty"`
}

// TravelStoryTheme represents the theme configuration.
type TravelStoryTheme struct {
	ColorScheme string `json:"colorScheme,omitempty"`
	FontFamily  string `json:"fontFamily,omitempty"`
	LayoutStyle string `json:"layoutStyle,omitempty"`
}

// TravelStory represents a curated travel journal page.
type TravelStory struct {
	// ID is the system generated unique identifier.
	ID int32
	// UID is the user defined unique identifier.
	UID string

	// Standard fields
	CreatorID int32
	CreatedTs int64
	UpdatedTs int64

	// Domain specific fields
	Title       string
	Description string
	CoverImage  string
	StartDate   *int64
	EndDate     *int64
	Destination string
	Visibility  Visibility
	Payload     *TravelStoryPayload

	// Composed fields - memo IDs associated with this story
	MemoIDs []int32
}

// TravelStoryMemo represents the junction between travel story and memo.
type TravelStoryMemo struct {
	TravelStoryID int32
	MemoID        int32
	DisplayOrder  int32
}

// FindTravelStory specifies the conditions for finding travel stories.
type FindTravelStory struct {
	ID             *int32
	UID            *string
	CreatorID      *int32
	VisibilityList []Visibility

	// Pagination
	Limit  *int
	Offset *int
}

// UpdateTravelStory specifies the fields to update.
type UpdateTravelStory struct {
	ID          int32
	UID         *string
	UpdatedTs   *int64
	Title       *string
	Description *string
	CoverImage  *string
	StartDate   *int64
	EndDate     *int64
	Destination *string
	Visibility  *Visibility
	Payload     *TravelStoryPayload
}

// DeleteTravelStory specifies which travel story to delete.
type DeleteTravelStory struct {
	ID int32
}

func (s *Store) CreateTravelStory(ctx context.Context, create *TravelStory) (*TravelStory, error) {
	if !base.UIDMatcher.MatchString(create.UID) {
		return nil, errors.New("invalid uid")
	}
	return s.driver.CreateTravelStory(ctx, create)
}

func (s *Store) ListTravelStories(ctx context.Context, find *FindTravelStory) ([]*TravelStory, error) {
	return s.driver.ListTravelStories(ctx, find)
}

func (s *Store) GetTravelStory(ctx context.Context, find *FindTravelStory) (*TravelStory, error) {
	list, err := s.ListTravelStories(ctx, find)
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

func (s *Store) UpdateTravelStory(ctx context.Context, update *UpdateTravelStory) error {
	if update.UID != nil && !base.UIDMatcher.MatchString(*update.UID) {
		return errors.New("invalid uid")
	}
	return s.driver.UpdateTravelStory(ctx, update)
}

func (s *Store) DeleteTravelStory(ctx context.Context, delete *DeleteTravelStory) error {
	return s.driver.DeleteTravelStory(ctx, delete)
}

// TravelStoryMemo operations

func (s *Store) UpsertTravelStoryMemo(ctx context.Context, upsert *TravelStoryMemo) (*TravelStoryMemo, error) {
	return s.driver.UpsertTravelStoryMemo(ctx, upsert)
}

func (s *Store) ListTravelStoryMemos(ctx context.Context, travelStoryID int32) ([]*TravelStoryMemo, error) {
	return s.driver.ListTravelStoryMemos(ctx, travelStoryID)
}

func (s *Store) DeleteTravelStoryMemos(ctx context.Context, travelStoryID int32) error {
	return s.driver.DeleteTravelStoryMemos(ctx, travelStoryID)
}