-- +goose Up
-- +goose StatementBegin


-- Teams table
CREATE TABLE IF NOT EXISTS teams (
    team_name VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE teams IS 'Stores team information';
COMMENT ON COLUMN teams.team_name IS 'Unique team name identifier';
COMMENT ON COLUMN teams.created_at IS 'Timestamp when team was created';


-- Users table
CREATE TABLE IF NOT EXISTS users (
    user_id VARCHAR(255) PRIMARY KEY,
    username VARCHAR(255) NOT NULL,
    team_name VARCHAR(255) NOT NULL REFERENCES teams(team_name) ON DELETE CASCADE,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE users IS 'Stores user accounts and team memberships';
COMMENT ON COLUMN users.user_id IS 'Unique user identifier';
COMMENT ON COLUMN users.username IS 'User display name';
COMMENT ON COLUMN users.team_name IS 'Team the user belongs to';
COMMENT ON COLUMN users.is_active IS 'Whether user can be assigned as reviewer';
COMMENT ON COLUMN users.created_at IS 'Timestamp when user was created';
COMMENT ON COLUMN users.updated_at IS 'Timestamp when user was last updated';

CREATE INDEX IF NOT EXISTS idx_users_team_name ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_users_team_active ON users(team_name, is_active) WHERE is_active = true;


-- Pull requests table
CREATE TABLE IF NOT EXISTS pull_requests (
    pull_request_id VARCHAR(255) PRIMARY KEY,
    pull_request_name VARCHAR(255) NOT NULL,
    author_id VARCHAR(255) NOT NULL REFERENCES users(user_id),
    status VARCHAR(20) NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    merged_at TIMESTAMP
);

COMMENT ON TABLE pull_requests IS 'Stores pull request information';
COMMENT ON COLUMN pull_requests.pull_request_id IS 'Unique pull request identifier';
COMMENT ON COLUMN pull_requests.pull_request_name IS 'Title/name of the pull request';
COMMENT ON COLUMN pull_requests.author_id IS 'User who created the pull request';
COMMENT ON COLUMN pull_requests.status IS 'Current state: OPEN or MERGED';
COMMENT ON COLUMN pull_requests.created_at IS 'Timestamp when PR was created';
COMMENT ON COLUMN pull_requests.merged_at IS 'Timestamp when PR was merged (null if open)';


CREATE INDEX IF NOT EXISTS idx_pull_requests_author ON pull_requests(author_id);
CREATE INDEX IF NOT EXISTS idx_pull_requests_status ON pull_requests(status);


-- PR reviewers junction table (N:M)
CREATE TABLE IF NOT EXISTS pr_reviewers (
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL REFERENCES users(user_id),
    assigned_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (pull_request_id, user_id)
);

COMMENT ON TABLE pr_reviewers IS 'Junction table for PR-reviewer assignments';
COMMENT ON COLUMN pr_reviewers.pull_request_id IS 'Pull request being reviewed';
COMMENT ON COLUMN pr_reviewers.user_id IS 'User assigned as reviewer';
COMMENT ON COLUMN pr_reviewers.assigned_at IS 'Timestamp when reviewer was assigned';

CREATE INDEX IF NOT EXISTS idx_pr_reviewers_user_id ON pr_reviewers(user_id);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_pr_id ON pr_reviewers(pull_request_id);
-- +goose StatementEnd


-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
-- +goose StatementEnd
