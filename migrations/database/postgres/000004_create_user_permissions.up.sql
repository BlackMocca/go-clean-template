CREATE TABLE user_permissions (
    user_id UUID NOT NULL,
    permission_id UUID NOT NULL,

    CONSTRAINT pk_user_permissions
        PRIMARY KEY (user_id, permission_id),

    CONSTRAINT fk_user_permissions_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT fk_user_permissions_permission
        FOREIGN KEY (permission_id)
        REFERENCES permissions(id)
        ON DELETE CASCADE
);

CREATE INDEX idx_user_permissions_permission_id
    ON user_permissions (permission_id);

COMMENT ON TABLE user_permissions IS
    'Maps users to permissions';

COMMENT ON COLUMN user_permissions.user_id IS
    'Reference to the user';

COMMENT ON COLUMN user_permissions.permission_id IS
    'Reference to the permission';