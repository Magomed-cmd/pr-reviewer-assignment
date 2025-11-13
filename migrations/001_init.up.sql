CREATE TYPE pr_status_enum AS ENUM ('OPEN', 'MERGED');

CREATE TABLE teams (
    team_name VARCHAR PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE users (
    user_id VARCHAR PRIMARY KEY,
    username VARCHAR NOT NULL,
    team_name VARCHAR NOT NULL REFERENCES teams(team_name),
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE pull_requests (
    pull_request_id VARCHAR PRIMARY KEY,
    pull_request_name VARCHAR NOT NULL,
    author_id VARCHAR NOT NULL REFERENCES users(user_id),
    status pr_status_enum NOT NULL DEFAULT 'OPEN',
    created_at TIMESTAMP DEFAULT NOW(),
    merged_at TIMESTAMP NULL
);

CREATE TABLE pr_reviewers (
    pull_request_id VARCHAR REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    user_id VARCHAR REFERENCES users(user_id) ON DELETE CASCADE,
    assigned_at TIMESTAMP DEFAULT NOW(),

    PRIMARY KEY (pull_request_id, user_id)
);

CREATE INDEX idx_users_team ON users(team_name);
CREATE INDEX idx_users_active ON users(is_active);
CREATE INDEX idx_pr_author ON pull_requests(author_id);
CREATE INDEX idx_pr_status ON pull_requests(status);
CREATE INDEX idx_pr_reviewers_user ON pr_reviewers(user_id);
CREATE INDEX idx_pr_reviewers_pr ON pr_reviewers(pull_request_id);
