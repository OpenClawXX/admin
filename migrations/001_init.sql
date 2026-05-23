-- +goose Up
-- SQL in this section is executed when the migration is applied

CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    supervisor_id BIGINT,
    description TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    real_name VARCHAR(50) DEFAULT '',
    email VARCHAR(100) DEFAULT '',
    phone VARCHAR(20) DEFAULT '',
    role VARCHAR(20) NOT NULL DEFAULT 'engineer',
    team_id BIGINT REFERENCES teams(id),
    skills JSONB DEFAULT '[]',
    status INT DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS projects (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    description TEXT DEFAULT '',
    status VARCHAR(20) DEFAULT 'active',
    manager_id BIGINT NOT NULL REFERENCES users(id),
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS assets (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    type VARCHAR(50) NOT NULL,
    ip VARCHAR(50) DEFAULT '',
    status VARCHAR(20) DEFAULT 'active',
    location VARCHAR(200) DEFAULT '',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS tickets (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    description TEXT DEFAULT '',
    type VARCHAR(20) NOT NULL,
    priority VARCHAR(10) NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'created',
    creator_id BIGINT NOT NULL REFERENCES users(id),
    assignee_id BIGINT REFERENCES users(id),
    project_id BIGINT REFERENCES projects(id),
    asset_id BIGINT REFERENCES assets(id),
    alert_id BIGINT,
    sla_deadline TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tickets_status ON tickets(status);
CREATE INDEX idx_tickets_assignee ON tickets(assignee_id);
CREATE INDEX idx_tickets_creator ON tickets(creator_id);
CREATE INDEX idx_tickets_project ON tickets(project_id);
CREATE INDEX idx_tickets_priority ON tickets(priority);
CREATE INDEX idx_tickets_created ON tickets(created_at);

CREATE TABLE IF NOT EXISTS ticket_logs (
    id BIGSERIAL PRIMARY KEY,
    ticket_id BIGINT NOT NULL REFERENCES tickets(id),
    action VARCHAR(50) NOT NULL,
    operator_id BIGINT NOT NULL REFERENCES users(id),
    content TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_ticket_logs_ticket ON ticket_logs(ticket_id);

CREATE TABLE IF NOT EXISTS scores (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    ticket_id BIGINT REFERENCES tickets(id),
    period VARCHAR(20) NOT NULL,
    dimension VARCHAR(50) NOT NULL,
    score NUMERIC(5,2) DEFAULT 0,
    remark TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_scores_user_period ON scores(user_id, period);

CREATE TABLE IF NOT EXISTS knowledge_articles (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(500) NOT NULL,
    content TEXT NOT NULL,
    tags JSONB DEFAULT '[]',
    author_id BIGINT NOT NULL REFERENCES users(id),
    related_tickets JSONB DEFAULT '[]',
    view_count INT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS schedules (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    date VARCHAR(10) NOT NULL,
    shift_type VARCHAR(20) NOT NULL,
    remark TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_schedules_user_date ON schedules(user_id, date);

CREATE TABLE IF NOT EXISTS monitor_alerts (
    id BIGSERIAL PRIMARY KEY,
    source VARCHAR(50) NOT NULL,
    alert_name VARCHAR(200) NOT NULL,
    level VARCHAR(20) NOT NULL,
    ticket_id BIGINT REFERENCES tickets(id),
    raw_payload JSONB DEFAULT '{}',
    status VARCHAR(20) DEFAULT 'pending',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS notifications (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    title VARCHAR(200) NOT NULL,
    content TEXT DEFAULT '',
    type VARCHAR(20) DEFAULT 'info',
    is_read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_notifications_user ON notifications(user_id, is_read);

CREATE TABLE IF NOT EXISTS operation_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    resource VARCHAR(50) NOT NULL,
    resource_id BIGINT DEFAULT 0,
    detail TEXT DEFAULT '',
    ip VARCHAR(50) DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_operation_logs_user ON operation_logs(user_id);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back
DROP TABLE IF EXISTS operation_logs;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS monitor_alerts;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS knowledge_articles;
DROP TABLE IF EXISTS scores;
DROP TABLE IF EXISTS ticket_logs;
DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
