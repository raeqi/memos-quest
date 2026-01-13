package travelstory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lithammer/shortuuid/v4"

	"github.com/usememos/memos/internal/profile"
	"github.com/usememos/memos/server/auth"
	"github.com/usememos/memos/store"
)

// TravelStoryService provides REST API endpoints for travel stories.
type TravelStoryService struct {
	Profile *profile.Profile
	Store   *store.Store
	Secret  string
}

// NewTravelStoryService creates a new TravelStoryService.
func NewTravelStoryService(secret string, profile *profile.Profile, store *store.Store) *TravelStoryService {
	return &TravelStoryService{
		Secret:  secret,
		Profile: profile,
		Store:   store,
	}
}

// Register registers the travel story routes with the Echo server.
func (s *TravelStoryService) Register(e *echo.Echo) {
	g := e.Group("/api/v1/travelStories")

	g.POST("", s.CreateTravelStory)
	g.GET("", s.ListTravelStories)
	g.GET("/:uid", s.GetTravelStory)
	g.PATCH("/:uid", s.UpdateTravelStory)
	g.DELETE("/:uid", s.DeleteTravelStory)
	g.POST("/detectTrips", s.DetectTrips)
	g.POST("/generateContent", s.GenerateStoryContent)
	g.GET("/:uid/export", s.ExportTravelStory)
	g.GET("/:uid/view", s.ViewTravelStory)
}

// API Request/Response types

type CreateTravelStoryRequest struct {
	Title               string       `json:"title"`
	Description         string       `json:"description"`
	MemoNames           []string     `json:"memoNames"`
	Theme               *StoryTheme  `json:"theme"`
	Visibility          string       `json:"visibility"`
	AutoGenerateContent bool         `json:"autoGenerateContent"`
}

type TravelStoryResponse struct {
	Name        string          `json:"name"`
	Creator     string          `json:"creator"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	CoverImage  string          `json:"coverImage"`
	StartDate   *int64          `json:"startDate,omitempty"`
	EndDate     *int64          `json:"endDate,omitempty"`
	Destination string          `json:"destination"`
	Tags        []string        `json:"tags"`
	Sections    []StorySection  `json:"sections"`
	Theme       *StoryTheme     `json:"theme"`
	Visibility  string          `json:"visibility"`
	CreateTime  int64           `json:"createTime"`
	UpdateTime  int64           `json:"updateTime"`
	SourceMemos []string        `json:"sourceMemos"`
}

type StorySection struct {
	Title      string       `json:"title"`
	Content    string       `json:"content"`
	Date       *int64       `json:"date,omitempty"`
	Location   string       `json:"location"`
	Images     []StoryImage `json:"images"`
	SourceMemo string       `json:"sourceMemo"`
}

type StoryImage struct {
	URL          string `json:"url"`
	Caption      string `json:"caption"`
	AltText      string `json:"altText"`
	QualityScore int32  `json:"qualityScore"`
	Selected     bool   `json:"selected"`
}

type StoryTheme struct {
	ColorScheme string `json:"colorScheme"`
	FontFamily  string `json:"fontFamily"`
	LayoutStyle string `json:"layoutStyle"`
}

type DetectedTrip struct {
	SuggestedTitle string   `json:"suggestedTitle"`
	Destination    string   `json:"destination"`
	StartDate      int64    `json:"startDate"`
	EndDate        int64    `json:"endDate"`
	Tags           []string `json:"tags"`
	MemoNames      []string `json:"memoNames"`
	Confidence     int32    `json:"confidence"`
}

type DetectTripsRequest struct {
	Tags         []string `json:"tags"`
	StartDate    *int64   `json:"startDate"`
	EndDate      *int64   `json:"endDate"`
	MinMemoCount int      `json:"minMemoCount"`
}

type GenerateContentRequest struct {
	MemoNames    []string `json:"memoNames"`
	TitleContext string   `json:"titleContext"`
}

type GenerateContentResponse struct {
	Sections            []StorySection `json:"sections"`
	SuggestedCoverImage string         `json:"suggestedCoverImage"`
	SuggestedTitle      string         `json:"suggestedTitle"`
	DetectedDestination string         `json:"detectedDestination"`
}

// Helper to get current user from context
func (s *TravelStoryService) getCurrentUser(c echo.Context) (*store.User, error) {
	ctx := c.Request().Context()

	// Check for user claims (JWT)
	claims := auth.GetUserClaims(ctx)
	if claims != nil {
		user, err := s.Store.GetUser(ctx, &store.FindUser{ID: &claims.UserID})
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	// Check for user ID in context (PAT)
	userID := auth.GetUserID(ctx)
	if userID != 0 {
		user, err := s.Store.GetUser(ctx, &store.FindUser{ID: &userID})
		if err != nil {
			return nil, err
		}
		return user, nil
	}

	return nil, nil
}

// CreateTravelStory creates a new travel story from selected memos.
func (s *TravelStoryService) CreateTravelStory(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := s.getCurrentUser(c)
	if err != nil || user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	var req CreateTravelStoryRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if req.Title == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "title is required")
	}
	if len(req.MemoNames) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "at least one memo is required")
	}

	// Convert memo names to IDs and validate ownership
	memoIDs := []int32{}
	for _, name := range req.MemoNames {
		uid := strings.TrimPrefix(name, "memos/")
		memo, err := s.Store.GetMemo(ctx, &store.FindMemo{UID: &uid})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to get memo")
		}
		if memo == nil {
			return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("memo not found: %s", name))
		}
		if memo.CreatorID != user.ID {
			return echo.NewHTTPError(http.StatusForbidden, "can only create stories from your own memos")
		}
		memoIDs = append(memoIDs, memo.ID)
	}

	// Set default visibility
	visibility := store.Private
	if req.Visibility == "PUBLIC" {
		visibility = store.Public
	} else if req.Visibility == "PROTECTED" {
		visibility = store.Protected
	}

	// Create payload
	payload := &store.TravelStoryPayload{}
	if req.Theme != nil {
		payload.Theme = &store.TravelStoryTheme{
			ColorScheme: req.Theme.ColorScheme,
			FontFamily:  req.Theme.FontFamily,
			LayoutStyle: req.Theme.LayoutStyle,
		}
	}

	// Generate content if requested
	if req.AutoGenerateContent {
		sections, coverImage, destination := s.generateContentFromMemos(ctx, memoIDs)
		for _, section := range sections {
			payload.Sections = append(payload.Sections, &store.TravelStorySection{
				Title:         section.Title,
				Content:       section.Content,
				Date:          ptrToInt64(section.Date),
				Location:      section.Location,
				SourceMemoUID: strings.TrimPrefix(section.SourceMemo, "memos/"),
			})
		}
		if coverImage != "" {
			req.Description = destination // Use destination as description if empty
		}
	}

	// Create the travel story
	create := &store.TravelStory{
		UID:         shortuuid.New(),
		CreatorID:   user.ID,
		Title:       req.Title,
		Description: req.Description,
		Visibility:  visibility,
		Payload:     payload,
		MemoIDs:     memoIDs,
	}

	story, err := s.Store.CreateTravelStory(ctx, create)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create travel story")
	}

	return c.JSON(http.StatusOK, s.convertToResponse(story, user.ID))
}

// ListTravelStories lists travel stories for the current user.
func (s *TravelStoryService) ListTravelStories(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := s.getCurrentUser(c)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get user")
	}

	find := &store.FindTravelStory{}
	if user != nil {
		// Show user's own stories plus public/protected
		find.CreatorID = &user.ID
	} else {
		// Only show public stories for unauthenticated users
		find.VisibilityList = []store.Visibility{store.Public}
	}

	stories, err := s.Store.ListTravelStories(ctx, find)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list travel stories")
	}

	response := []TravelStoryResponse{}
	for _, story := range stories {
		response = append(response, s.convertToResponse(story, story.CreatorID))
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"travelStories": response,
	})
}

// GetTravelStory gets a travel story by UID.
func (s *TravelStoryService) GetTravelStory(c echo.Context) error {
	ctx := c.Request().Context()
	uid := c.Param("uid")

	story, err := s.Store.GetTravelStory(ctx, &store.FindTravelStory{UID: &uid})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get travel story")
	}
	if story == nil {
		return echo.NewHTTPError(http.StatusNotFound, "travel story not found")
	}

	// Check visibility
	user, _ := s.getCurrentUser(c)
	if story.Visibility != store.Public {
		if user == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
		}
		if story.Visibility == store.Private && story.CreatorID != user.ID {
			return echo.NewHTTPError(http.StatusForbidden, "access denied")
		}
	}

	return c.JSON(http.StatusOK, s.convertToResponse(story, story.CreatorID))
}

// UpdateTravelStory updates a travel story.
func (s *TravelStoryService) UpdateTravelStory(c echo.Context) error {
	ctx := c.Request().Context()
	uid := c.Param("uid")

	user, err := s.getCurrentUser(c)
	if err != nil || user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	story, err := s.Store.GetTravelStory(ctx, &store.FindTravelStory{UID: &uid})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get travel story")
	}
	if story == nil {
		return echo.NewHTTPError(http.StatusNotFound, "travel story not found")
	}
	if story.CreatorID != user.ID {
		return echo.NewHTTPError(http.StatusForbidden, "can only update your own stories")
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(c.Request().Body).Decode(&updates); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	update := &store.UpdateTravelStory{ID: story.ID}
	now := time.Now().Unix()
	update.UpdatedTs = &now

	if title, ok := updates["title"].(string); ok {
		update.Title = &title
	}
	if desc, ok := updates["description"].(string); ok {
		update.Description = &desc
	}
	if cover, ok := updates["coverImage"].(string); ok {
		update.CoverImage = &cover
	}
	if dest, ok := updates["destination"].(string); ok {
		update.Destination = &dest
	}
	if vis, ok := updates["visibility"].(string); ok {
		var visibility store.Visibility
		switch vis {
		case "PUBLIC":
			visibility = store.Public
		case "PROTECTED":
			visibility = store.Protected
		default:
			visibility = store.Private
		}
		update.Visibility = &visibility
	}

	if err := s.Store.UpdateTravelStory(ctx, update); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to update travel story")
	}

	// Fetch updated story
	story, _ = s.Store.GetTravelStory(ctx, &store.FindTravelStory{ID: &story.ID})
	return c.JSON(http.StatusOK, s.convertToResponse(story, user.ID))
}

// DeleteTravelStory deletes a travel story.
func (s *TravelStoryService) DeleteTravelStory(c echo.Context) error {
	ctx := c.Request().Context()
	uid := c.Param("uid")

	user, err := s.getCurrentUser(c)
	if err != nil || user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	story, err := s.Store.GetTravelStory(ctx, &store.FindTravelStory{UID: &uid})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get travel story")
	}
	if story == nil {
		return echo.NewHTTPError(http.StatusNotFound, "travel story not found")
	}
	if story.CreatorID != user.ID {
		return echo.NewHTTPError(http.StatusForbidden, "can only delete your own stories")
	}

	if err := s.Store.DeleteTravelStory(ctx, &store.DeleteTravelStory{ID: story.ID}); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to delete travel story")
	}

	return c.NoContent(http.StatusNoContent)
}

// DetectTrips analyzes memos to detect potential trips.
func (s *TravelStoryService) DetectTrips(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := s.getCurrentUser(c)
	if err != nil || user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	var req DetectTripsRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	// Get user's memos
	find := &store.FindMemo{
		CreatorID: &user.ID,
	}
	memos, err := s.Store.ListMemos(ctx, find)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to list memos")
	}

	// Detect trips using tag clustering and date proximity
	trips := s.detectTripsFromMemos(memos, req)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"trips": trips,
	})
}

// GenerateStoryContent generates refined content from memos.
func (s *TravelStoryService) GenerateStoryContent(c echo.Context) error {
	ctx := c.Request().Context()

	user, err := s.getCurrentUser(c)
	if err != nil || user == nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
	}

	var req GenerateContentRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	// Convert memo names to IDs
	memoIDs := []int32{}
	for _, name := range req.MemoNames {
		uid := strings.TrimPrefix(name, "memos/")
		memo, err := s.Store.GetMemo(ctx, &store.FindMemo{UID: &uid})
		if err != nil || memo == nil {
			continue
		}
		if memo.CreatorID != user.ID {
			continue
		}
		memoIDs = append(memoIDs, memo.ID)
	}

	sections, coverImage, destination := s.generateContentFromMemos(ctx, memoIDs)

	// Generate suggested title
	suggestedTitle := req.TitleContext
	if suggestedTitle == "" && destination != "" {
		suggestedTitle = fmt.Sprintf("My Trip to %s", destination)
	}

	return c.JSON(http.StatusOK, GenerateContentResponse{
		Sections:            sections,
		SuggestedCoverImage: coverImage,
		SuggestedTitle:      suggestedTitle,
		DetectedDestination: destination,
	})
}

// ExportTravelStory exports a travel story as HTML.
func (s *TravelStoryService) ExportTravelStory(c echo.Context) error {
	ctx := c.Request().Context()
	uid := c.Param("uid")
	format := c.QueryParam("format")
	if format == "" {
		format = "html"
	}

	story, err := s.Store.GetTravelStory(ctx, &store.FindTravelStory{UID: &uid})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get travel story")
	}
	if story == nil {
		return echo.NewHTTPError(http.StatusNotFound, "travel story not found")
	}

	// Check visibility
	user, _ := s.getCurrentUser(c)
	if story.Visibility != store.Public {
		if user == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
		}
		if story.Visibility == store.Private && story.CreatorID != user.ID {
			return echo.NewHTTPError(http.StatusForbidden, "access denied")
		}
	}

	html := s.generateHTMLExport(story)

	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Response().Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.html\"", story.Title))
	return c.String(http.StatusOK, html)
}

// ViewTravelStory renders the shareable travel story page.
func (s *TravelStoryService) ViewTravelStory(c echo.Context) error {
	ctx := c.Request().Context()
	uid := c.Param("uid")

	story, err := s.Store.GetTravelStory(ctx, &store.FindTravelStory{UID: &uid})
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get travel story")
	}
	if story == nil {
		return echo.NewHTTPError(http.StatusNotFound, "travel story not found")
	}

	// Check visibility
	user, _ := s.getCurrentUser(c)
	if story.Visibility != store.Public {
		if user == nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
		}
		if story.Visibility == store.Private && story.CreatorID != user.ID {
			return echo.NewHTTPError(http.StatusForbidden, "access denied")
		}
	}

	html := s.generateViewPage(story)
	c.Response().Header().Set("Content-Type", "text/html; charset=utf-8")
	return c.String(http.StatusOK, html)
}

// Helper methods

func (s *TravelStoryService) convertToResponse(story *store.TravelStory, creatorID int32) TravelStoryResponse {
	resp := TravelStoryResponse{
		Name:        fmt.Sprintf("travelStories/%s", story.UID),
		Creator:     fmt.Sprintf("users/%d", creatorID),
		Title:       story.Title,
		Description: story.Description,
		CoverImage:  story.CoverImage,
		StartDate:   story.StartDate,
		EndDate:     story.EndDate,
		Destination: story.Destination,
		Visibility:  string(story.Visibility),
		CreateTime:  story.CreatedTs,
		UpdateTime:  story.UpdatedTs,
	}

	if story.Payload != nil {
		resp.Tags = story.Payload.Tags
		if story.Payload.Theme != nil {
			resp.Theme = &StoryTheme{
				ColorScheme: story.Payload.Theme.ColorScheme,
				FontFamily:  story.Payload.Theme.FontFamily,
				LayoutStyle: story.Payload.Theme.LayoutStyle,
			}
		}
		for _, section := range story.Payload.Sections {
			s := StorySection{
				Title:      section.Title,
				Content:    section.Content,
				Location:   section.Location,
				SourceMemo: fmt.Sprintf("memos/%s", section.SourceMemoUID),
			}
			if section.Date != 0 {
				d := section.Date
				s.Date = &d
			}
			for _, img := range section.Images {
				s.Images = append(s.Images, StoryImage{
					URL:          img.URL,
					Caption:      img.Caption,
					AltText:      img.AltText,
					QualityScore: img.QualityScore,
					Selected:     img.Selected,
				})
			}
			resp.Sections = append(resp.Sections, s)
		}
	}

	for _, memoID := range story.MemoIDs {
		resp.SourceMemos = append(resp.SourceMemos, fmt.Sprintf("memos/%d", memoID))
	}

	return resp
}

func (s *TravelStoryService) detectTripsFromMemos(memos []*store.Memo, req DetectTripsRequest) []DetectedTrip {
	trips := []DetectedTrip{}

	// Group memos by common tags
	tagGroups := make(map[string][]*store.Memo)
	travelTags := []string{"travel", "trip", "vacation", "holiday", "journey"}

	for _, memo := range memos {
		if memo.Payload == nil {
			continue
		}
		for _, tag := range memo.Payload.Tags {
			tagLower := strings.ToLower(tag)
			// Check if it's a travel-related tag or location tag
			isTravelTag := false
			for _, tt := range travelTags {
				if strings.Contains(tagLower, tt) {
					isTravelTag = true
					break
				}
			}
			if isTravelTag || isLocationTag(tag) {
				tagGroups[tag] = append(tagGroups[tag], memo)
			}
		}
	}

	// Convert tag groups to trips
	for tag, groupMemos := range tagGroups {
		if len(groupMemos) < max(req.MinMemoCount, 2) {
			continue
		}

		// Sort by date
		sort.Slice(groupMemos, func(i, j int) bool {
			return groupMemos[i].CreatedTs < groupMemos[j].CreatedTs
		})

		// Check date proximity (within 30 days)
		startDate := groupMemos[0].CreatedTs
		endDate := groupMemos[len(groupMemos)-1].CreatedTs
		if endDate-startDate > 30*24*60*60 {
			continue // Too spread out
		}

		// Collect memo names
		memoNames := []string{}
		commonTags := make(map[string]int)
		for _, m := range groupMemos {
			memoNames = append(memoNames, fmt.Sprintf("memos/%s", m.UID))
			if m.Payload != nil {
				for _, t := range m.Payload.Tags {
					commonTags[t]++
				}
			}
		}

		// Find most common tags
		topTags := []string{}
		for t, count := range commonTags {
			if count >= len(groupMemos)/2 {
				topTags = append(topTags, t)
			}
		}

		trip := DetectedTrip{
			SuggestedTitle: fmt.Sprintf("Trip: %s", tag),
			Destination:    extractDestination(tag),
			StartDate:      startDate,
			EndDate:        endDate,
			Tags:           topTags,
			MemoNames:      memoNames,
			Confidence:     int32(min(100, len(groupMemos)*20)),
		}
		trips = append(trips, trip)
	}

	return trips
}

func (s *TravelStoryService) generateContentFromMemos(ctx context.Context, memoIDs []int32) ([]StorySection, string, string) {
	sections := []StorySection{}
	var coverImage string
	var destination string

	for _, memoID := range memoIDs {
		id := memoID
		memo, err := s.Store.GetMemo(ctx, &store.FindMemo{ID: &id})
		if err != nil || memo == nil {
			continue
		}

		// Get attachments for this memo
		attachments, _ := s.Store.ListAttachments(ctx, &store.FindAttachment{MemoID: &memoID})
		images := []StoryImage{}
		for _, att := range attachments {
			if strings.HasPrefix(att.Type, "image/") {
				img := StoryImage{
					URL:          fmt.Sprintf("/file/attachments/%s", att.UID),
					QualityScore: 80, // Default score
					Selected:     true,
				}
				images = append(images, img)
				if coverImage == "" {
					coverImage = img.URL
				}
			}
		}

		// Extract location from memo payload
		location := ""
		if memo.Payload != nil && memo.Payload.Location != nil {
			location = memo.Payload.Location.Placeholder
			if destination == "" && location != "" {
				destination = location
			}
		}

		// Refine content (keep authentic voice but clean up)
		content := refineContent(memo.Content)

		section := StorySection{
			Title:      extractTitle(memo.Content),
			Content:    content,
			Date:       &memo.CreatedTs,
			Location:   location,
			Images:     images,
			SourceMemo: fmt.Sprintf("memos/%s", memo.UID),
		}
		sections = append(sections, section)
	}

	return sections, coverImage, destination
}

func (s *TravelStoryService) generateHTMLExport(story *store.TravelStory) string {
	theme := getThemeCSS(story.Payload)

	sectionsHTML := ""
	if story.Payload != nil {
		for _, section := range story.Payload.Sections {
			imagesHTML := ""
			for _, img := range section.Images {
				if img.Selected {
					imagesHTML += fmt.Sprintf(`<img src="%s" alt="%s" class="story-image" loading="lazy">`, img.URL, img.AltText)
				}
			}

			dateStr := ""
			if section.Date != 0 {
				dateStr = time.Unix(section.Date, 0).Format("January 2, 2006")
			}

			sectionsHTML += fmt.Sprintf(`
				<section class="story-section">
					<h2>%s</h2>
					%s
					<div class="section-meta">
						<span class="date">%s</span>
						<span class="location">%s</span>
					</div>
					<div class="section-content">%s</div>
					<div class="section-images">%s</div>
				</section>
			`, section.Title, "", dateStr, section.Location, section.Content, imagesHTML)
		}
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s</title>
	<style>
		%s
		* { box-sizing: border-box; margin: 0; padding: 0; }
		body { font-family: var(--font-family); line-height: 1.6; color: var(--text-color); background: var(--bg-color); }
		.story-container { max-width: 900px; margin: 0 auto; padding: 2rem; }
		.story-header { text-align: center; margin-bottom: 3rem; padding: 4rem 2rem; background: var(--accent-color); color: white; border-radius: 12px; }
		.story-header h1 { font-size: 2.5rem; margin-bottom: 0.5rem; }
		.story-header .description { font-size: 1.2rem; opacity: 0.9; }
		.story-header .dates { margin-top: 1rem; font-size: 0.9rem; }
		.story-section { margin-bottom: 3rem; padding: 2rem; background: white; border-radius: 8px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
		.story-section h2 { color: var(--accent-color); margin-bottom: 1rem; }
		.section-meta { display: flex; gap: 1rem; color: #666; font-size: 0.9rem; margin-bottom: 1rem; }
		.section-content { white-space: pre-wrap; }
		.section-images { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 1rem; margin-top: 1.5rem; }
		.story-image { width: 100%%; height: 200px; object-fit: cover; border-radius: 8px; }
		@media (max-width: 640px) {
			.story-header h1 { font-size: 1.8rem; }
			.story-container { padding: 1rem; }
			.story-section { padding: 1.5rem; }
		}
	</style>
</head>
<body>
	<div class="story-container">
		<header class="story-header">
			<h1>%s</h1>
			<p class="description">%s</p>
			<p class="dates">%s</p>
		</header>
		<main class="story-content">
			%s
		</main>
	</div>
</body>
</html>`, story.Title, theme, story.Title, story.Description, story.Destination, sectionsHTML)
}

func (s *TravelStoryService) generateViewPage(story *store.TravelStory) string {
	return s.generateHTMLExport(story)
}

// Utility functions

func isLocationTag(tag string) bool {
	// Common location indicators
	locations := []string{"city", "country", "state", "beach", "mountain", "island", "park"}
	tagLower := strings.ToLower(tag)
	for _, loc := range locations {
		if strings.Contains(tagLower, loc) {
			return true
		}
	}
	// Check if tag starts with capital (likely a place name)
	if len(tag) > 0 && tag[0] >= 'A' && tag[0] <= 'Z' {
		return true
	}
	return false
}

func extractDestination(tag string) string {
	// Remove common prefixes
	tag = strings.TrimPrefix(tag, "#")
	tag = strings.TrimPrefix(tag, "travel-")
	tag = strings.TrimPrefix(tag, "trip-")
	// Capitalize first letter
	if len(tag) > 0 {
		return strings.ToUpper(string(tag[0])) + tag[1:]
	}
	return tag
}

func extractTitle(content string) string {
	// Get first line or first sentence
	lines := strings.Split(content, "\n")
	if len(lines) > 0 {
		firstLine := strings.TrimSpace(lines[0])
		// Remove markdown headers
		firstLine = strings.TrimLeft(firstLine, "# ")
		if len(firstLine) > 50 {
			firstLine = firstLine[:47] + "..."
		}
		if firstLine != "" {
			return firstLine
		}
	}
	return "Untitled"
}

func refineContent(content string) string {
	// Remove markdown headers (keep content clean)
	lines := strings.Split(content, "\n")
	refined := []string{}
	for _, line := range lines {
		// Skip empty lines at start
		if len(refined) == 0 && strings.TrimSpace(line) == "" {
			continue
		}
		// Remove header markers but keep text
		line = strings.TrimLeft(line, "# ")
		refined = append(refined, line)
	}
	return strings.Join(refined, "\n")
}

func getThemeCSS(payload *store.TravelStoryPayload) string {
	colorScheme := "ocean"
	fontFamily := "Georgia, serif"

	if payload != nil && payload.Theme != nil {
		if payload.Theme.ColorScheme != "" {
			colorScheme = payload.Theme.ColorScheme
		}
		if payload.Theme.FontFamily != "" {
			fontFamily = payload.Theme.FontFamily
		}
	}

	colors := map[string]struct{ accent, bg, text string }{
		"ocean":   {"#0077b6", "#f0f9ff", "#1e3a5f"},
		"sunset":  {"#e85d04", "#fff8f0", "#4a2c0a"},
		"forest":  {"#2d6a4f", "#f0fff4", "#1b4332"},
		"minimal": {"#333333", "#ffffff", "#1a1a1a"},
	}

	c := colors["ocean"]
	if theme, ok := colors[colorScheme]; ok {
		c = theme
	}

	return fmt.Sprintf(`
		:root {
			--accent-color: %s;
			--bg-color: %s;
			--text-color: %s;
			--font-family: %s;
		}
	`, c.accent, c.bg, c.text, fontFamily)
}

func ptrToInt64(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}