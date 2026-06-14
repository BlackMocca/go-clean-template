CREATE TABLE permissions (
    id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,

    CONSTRAINT pk_permissions PRIMARY KEY (id)
);

-- Indexes
CREATE UNIQUE INDEX uk_permissions_name
    ON permissions (name)
    WHERE deleted_at IS NULL;

CREATE INDEX idx_permissions_deleted_at
    ON permissions (deleted_at);

-- Comments
COMMENT ON TABLE permissions IS
    'Stores system permissions used for authorization and access control';

COMMENT ON COLUMN permissions.id IS
    'Unique identifier of the permission';

COMMENT ON COLUMN permissions.name IS
    'Permission name or permission code';

COMMENT ON COLUMN permissions.created_at IS
    'Timestamp when the permission was created';

COMMENT ON COLUMN permissions.updated_at IS
    'Timestamp when the permission was last updated';

COMMENT ON COLUMN permissions.deleted_at IS
    'Soft delete timestamp, NULL means active';