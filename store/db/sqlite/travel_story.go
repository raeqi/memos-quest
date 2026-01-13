package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	"github.com/usememos/memos/store"
)

func (d *DB) CreateTravelStory(ctx context.Context, create *store.TravelStory) (*store.TravelStory, error) {
	payloadBytes := []byte("{}")
	if create.Payload != nil {
		var err error
		payloadBytes, err = json.Marshal(create.Payload)
		if err != nil {
			return nil, errors.Wrap(err, "failed to marshal payload")
		}
	}

	fields := []string{"`uid`", "`creator_id`", "`title`", "`description`", "`cover_image`", "`destination`", "`visibility`", "`payload`"}
	placeholder := []string{"?", "?", "?", "?", "?", "?", "?", "?"}
	args := []any{create.UID, create.CreatorID, create.Title, create.Description, create.CoverImage, create.Destination, create.Visibility, string(payloadBytes)}

	if create.StartDate != nil {
		fields = append(fields, "`start_date`")
		placeholder = append(placeholder, "?")
		args = append(args, *create.StartDate)
	}
	if create.EndDate != nil {
		fields = append(fields, "`end_date`")
		placeholder = append(placeholder, "?")
		args = append(args, *create.EndDate)
	}

	stmt := "INSERT INTO `travel_story` (" + strings.Join(fields, ", ") + ") VALUES (" + strings.Join(placeholder, ", ") + ") RETURNING `id`, `created_ts`, `updated_ts`"
	if err := d.db.QueryRowContext(ctx, stmt, args...).Scan(
		&create.ID,
		&create.CreatedTs,
		&create.UpdatedTs,
	); err != nil {
		return nil, errors.Wrap(err, "failed to create travel story")
	}

	// Insert memo associations
	for i, memoID := range create.MemoIDs {
		_, err := d.UpsertTravelStoryMemo(ctx, &store.TravelStoryMemo{
			TravelStoryID: create.ID,
			MemoID:        memoID,
			DisplayOrder:  int32(i),
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to create travel story memo association")
		}
	}

	return create, nil
}

func (d *DB) ListTravelStories(ctx context.Context, find *store.FindTravelStory) ([]*store.TravelStory, error) {
	where, args := []string{"1 = 1"}, []any{}

	if find.ID != nil {
		where, args = append(where, "`id` = ?"), append(args, *find.ID)
	}
	if find.UID != nil {
		where, args = append(where, "`uid` = ?"), append(args, *find.UID)
	}
	if find.CreatorID != nil {
		where, args = append(where, "`creator_id` = ?"), append(args, *find.CreatorID)
	}
	if len(find.VisibilityList) > 0 {
		placeholders := make([]string, len(find.VisibilityList))
		for i, v := range find.VisibilityList {
			placeholders[i] = "?"
			args = append(args, v)
		}
		where = append(where, "`visibility` IN ("+strings.Join(placeholders, ",")+")")
	}

	query := `
		SELECT
			id,
			uid,
			creator_id,
			created_ts,
			updated_ts,
			title,
			description,
			cover_image,
			start_date,
			end_date,
			destination,
			visibility,
			payload
		FROM travel_story
		WHERE ` + strings.Join(where, " AND ") + `
		ORDER BY created_ts DESC`

	if find.Limit != nil {
		query += " LIMIT ?"
		args = append(args, *find.Limit)
	}
	if find.Offset != nil {
		query += " OFFSET ?"
		args = append(args, *find.Offset)
	}

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query travel stories")
	}
	defer rows.Close()

	list := []*store.TravelStory{}
	for rows.Next() {
		story := &store.TravelStory{}
		var payloadStr string
		var startDate, endDate sql.NullInt64

		if err := rows.Scan(
			&story.ID,
			&story.UID,
			&story.CreatorID,
			&story.CreatedTs,
			&story.UpdatedTs,
			&story.Title,
			&story.Description,
			&story.CoverImage,
			&startDate,
			&endDate,
			&story.Destination,
			&story.Visibility,
			&payloadStr,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan travel story")
		}

		if startDate.Valid {
			story.StartDate = &startDate.Int64
		}
		if endDate.Valid {
			story.EndDate = &endDate.Int64
		}

		payload := &store.TravelStoryPayload{}
		if payloadStr != "" && payloadStr != "{}" {
			if err := json.Unmarshal([]byte(payloadStr), payload); err != nil {
				return nil, errors.Wrap(err, "failed to unmarshal payload")
			}
		}
		story.Payload = payload

		// Load memo IDs
		memos, err := d.ListTravelStoryMemos(ctx, story.ID)
		if err != nil {
			return nil, errors.Wrap(err, "failed to list travel story memos")
		}
		for _, m := range memos {
			story.MemoIDs = append(story.MemoIDs, m.MemoID)
		}

		list = append(list, story)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "rows error")
	}

	return list, nil
}

func (d *DB) UpdateTravelStory(ctx context.Context, update *store.UpdateTravelStory) error {
	set, args := []string{}, []any{}

	if update.Title != nil {
		set, args = append(set, "`title` = ?"), append(args, *update.Title)
	}
	if update.Description != nil {
		set, args = append(set, "`description` = ?"), append(args, *update.Description)
	}
	if update.CoverImage != nil {
		set, args = append(set, "`cover_image` = ?"), append(args, *update.CoverImage)
	}
	if update.StartDate != nil {
		set, args = append(set, "`start_date` = ?"), append(args, *update.StartDate)
	}
	if update.EndDate != nil {
		set, args = append(set, "`end_date` = ?"), append(args, *update.EndDate)
	}
	if update.Destination != nil {
		set, args = append(set, "`destination` = ?"), append(args, *update.Destination)
	}
	if update.Visibility != nil {
		set, args = append(set, "`visibility` = ?"), append(args, *update.Visibility)
	}
	if update.Payload != nil {
		payloadBytes, err := json.Marshal(update.Payload)
		if err != nil {
			return errors.Wrap(err, "failed to marshal payload")
		}
		set, args = append(set, "`payload` = ?"), append(args, string(payloadBytes))
	}
	if update.UpdatedTs != nil {
		set, args = append(set, "`updated_ts` = ?"), append(args, *update.UpdatedTs)
	}

	if len(set) == 0 {
		return nil
	}

	args = append(args, update.ID)
	stmt := "UPDATE `travel_story` SET " + strings.Join(set, ", ") + " WHERE `id` = ?"
	_, err := d.db.ExecContext(ctx, stmt, args...)
	return err
}

func (d *DB) DeleteTravelStory(ctx context.Context, delete *store.DeleteTravelStory) error {
	// Delete memo associations first
	if err := d.DeleteTravelStoryMemos(ctx, delete.ID); err != nil {
		return errors.Wrap(err, "failed to delete travel story memos")
	}

	_, err := d.db.ExecContext(ctx, "DELETE FROM `travel_story` WHERE `id` = ?", delete.ID)
	return err
}

func (d *DB) UpsertTravelStoryMemo(ctx context.Context, upsert *store.TravelStoryMemo) (*store.TravelStoryMemo, error) {
	stmt := `
		INSERT INTO travel_story_memo (travel_story_id, memo_id, display_order)
		VALUES (?, ?, ?)
		ON CONFLICT(travel_story_id, memo_id) DO UPDATE SET display_order = excluded.display_order
	`
	_, err := d.db.ExecContext(ctx, stmt, upsert.TravelStoryID, upsert.MemoID, upsert.DisplayOrder)
	if err != nil {
		return nil, errors.Wrap(err, "failed to upsert travel story memo")
	}
	return upsert, nil
}

func (d *DB) ListTravelStoryMemos(ctx context.Context, travelStoryID int32) ([]*store.TravelStoryMemo, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT travel_story_id, memo_id, display_order
		FROM travel_story_memo
		WHERE travel_story_id = ?
		ORDER BY display_order ASC
	`, travelStoryID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query travel story memos")
	}
	defer rows.Close()

	list := []*store.TravelStoryMemo{}
	for rows.Next() {
		m := &store.TravelStoryMemo{}
		if err := rows.Scan(&m.TravelStoryID, &m.MemoID, &m.DisplayOrder); err != nil {
			return nil, errors.Wrap(err, "failed to scan travel story memo")
		}
		list = append(list, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (d *DB) DeleteTravelStoryMemos(ctx context.Context, travelStoryID int32) error {
	_, err := d.db.ExecContext(ctx, "DELETE FROM `travel_story_memo` WHERE `travel_story_id` = ?", travelStoryID)
	return err
}