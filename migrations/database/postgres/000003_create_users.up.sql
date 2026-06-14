CREATE TYPE user_role AS ENUM (
    'admin',
    'user',
    'moderator'
);

CREATE TABLE users (
    id UUID NOT NULL,

    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    role user_role NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT pk_users PRIMARY KEY (id)
);

CREATE INDEX idx_users_email_trgm
ON users
USING gin (email gin_trgm_ops);

CREATE INDEX idx_users_name_trgm
ON users
USING gin (name gin_trgm_ops);

-- Unique indexes
CREATE UNIQUE INDEX uk_users_email
    ON users (email)
    WHERE deleted_at IS NULL;

-- Query indexes
CREATE INDEX idx_users_role
    ON users (role);

CREATE INDEX idx_users_deleted_at
    ON users (deleted_at);

-- Comments
COMMENT ON TABLE users IS
    'Stores application user accounts and their roles';

COMMENT ON COLUMN users.id IS
    'Unique identifier of the user';

COMMENT ON COLUMN users.name IS
    'Display name of the user';

COMMENT ON COLUMN users.email IS
    'User email address used for communication and authentication';

COMMENT ON COLUMN users.role IS
    'Role assigned to the user, such as admin, user, or manager';

COMMENT ON COLUMN users.created_at IS
    'Timestamp when the user account was created';

COMMENT ON COLUMN users.updated_at IS
    'Timestamp when the user account was last updated';

COMMENT ON COLUMN users.deleted_at IS
    'Soft delete timestamp, NULL means active';