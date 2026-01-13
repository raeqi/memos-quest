-- travel_story table for storing curated travel journal pages
CREATE TABLE travel_story (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  uid TEXT NOT NULL UNIQUE,
  creator_id INTEGER NOT NULL,
  created_ts BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
  updated_ts BIGINT NOT NULL DEFAULT (strftime('%s', 'now')),
  title TEXT NOT NULL DEFAULT '',
  description TEXT NOT NULL DEFAULT '',
  cover_image TEXT NOT NULL DEFAULT '',
  start_date BIGINT,
  end_date BIGINT,
  destination TEXT NOT NULL DEFAULT '',
  visibility TEXT NOT NULL CHECK (visibility IN ('PUBLIC', 'PROTECTED', 'PRIVATE')) DEFAULT 'PRIVATE',
  payload TEXT NOT NULL DEFAULT '{}'
);

-- travel_story_memo junction table for tracking source memos
CREATE TABLE travel_story_memo (
  travel_story_id INTEGER NOT NULL,
  memo_id INTEGER NOT NULL,
  display_order INTEGER NOT NULL DEFAULT 0,
  UNIQUE(travel_story_id, memo_id)
);

-- Create indexes for efficient queries
CREATE INDEX idx_travel_story_creator_id ON travel_story(creator_id);
CREATE INDEX idx_travel_story_visibility ON travel_story(visibility);
CREATE INDEX idx_travel_story_memo_story_id ON travel_story_memo(travel_story_id);
CREATE INDEX idx_travel_story_memo_memo_id ON travel_story_memo(memo_id);
