-- travel_story table for storing curated travel journal pages
CREATE TABLE travel_story (
  id INT AUTO_INCREMENT PRIMARY KEY,
  uid VARCHAR(256) NOT NULL UNIQUE,
  creator_id INT NOT NULL,
  created_ts BIGINT NOT NULL DEFAULT (UNIX_TIMESTAMP()),
  updated_ts BIGINT NOT NULL DEFAULT (UNIX_TIMESTAMP()),
  title VARCHAR(512) NOT NULL DEFAULT '',
  description TEXT NOT NULL,
  cover_image TEXT NOT NULL,
  start_date BIGINT,
  end_date BIGINT,
  destination VARCHAR(512) NOT NULL DEFAULT '',
  visibility VARCHAR(32) NOT NULL DEFAULT 'PRIVATE',
  payload JSON NOT NULL
);

-- travel_story_memo junction table for tracking source memos
CREATE TABLE travel_story_memo (
  travel_story_id INT NOT NULL,
  memo_id INT NOT NULL,
  display_order INT NOT NULL DEFAULT 0,
  UNIQUE(travel_story_id, memo_id)
);

-- Create indexes for efficient queries
CREATE INDEX idx_travel_story_creator_id ON travel_story(creator_id);
CREATE INDEX idx_travel_story_visibility ON travel_story(visibility);
CREATE INDEX idx_travel_story_memo_story_id ON travel_story_memo(travel_story_id);
CREATE INDEX idx_travel_story_memo_memo_id ON travel_story_memo(memo_id);
