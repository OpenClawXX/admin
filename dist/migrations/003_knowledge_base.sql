CREATE TABLE IF NOT EXISTS kb_categories (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    parent_id BIGINT DEFAULT 0,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kb_articles (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    content TEXT DEFAULT '',
    content_html TEXT DEFAULT '',
    category_id BIGINT DEFAULT 0,
    status VARCHAR(20) DEFAULT 'draft',
    author_id BIGINT NOT NULL REFERENCES users(id),
    tags JSONB DEFAULT '[]',
    view_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS kb_files (
    id BIGSERIAL PRIMARY KEY,
    article_id BIGINT NOT NULL REFERENCES kb_articles(id) ON DELETE CASCADE,
    filename VARCHAR(500) NOT NULL,
    filepath VARCHAR(500) NOT NULL,
    filesize BIGINT DEFAULT 0,
    filetype VARCHAR(50) DEFAULT '',
    uploader_id BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kb_articles_category ON kb_articles(category_id);
CREATE INDEX IF NOT EXISTS idx_kb_articles_author ON kb_articles(author_id);
CREATE INDEX IF NOT EXISTS idx_kb_files_article ON kb_files(article_id);
