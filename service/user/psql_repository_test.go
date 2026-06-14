package user_test

import (
	"example/constants"
	"example/models"
	"example/service/user"
	"fmt"

	"testing"

	utilModels "github.com/BlackMocca/utils/models"
	"github.com/google/uuid"
	"github.com/guregu/null/v6/zero"
	"github.com/stretchr/testify/suite"
)

type UserRepositorySuite struct {
	PostgresRepositorySuite

	repository user.Repository
}

func TestUserRepositorySuite(t *testing.T) {
	suite.Run(
		t,
		new(UserRepositorySuite),
	)
}

func (s *UserRepositorySuite) SetupSuite() {
	s.PostgresRepositorySuite.SetupSuite()

	s.repository = user.NewPsqlRepository(s.client)
}

func (s *UserRepositorySuite) TestFindByID() {
	s.Run("success: returns user with permissions when found by ID", func() {
		// 1. Seed fixed UUIDs for permissions
		adminPermID := uuid.MustParse("a0000000-0000-0000-0000-000000000001")
		userPermID := uuid.MustParse("a0000000-0000-0000-0000-000000000002")

		// 2. Insert permissions (must exist before user_permissions FK)
		now := utilModels.NewTimestampFromNow()
		_, err := s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO permissions (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
			adminPermID, string(constants.RoleAdmin), now.ToPointer(), now.ToPointer(),
		)
		s.Require().NoError(err)
		_, err = s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO permissions (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
			userPermID, string(constants.RoleUser), now.ToPointer(), now.ToPointer(),
		)
		s.Require().NoError(err)

		// 3. Insert user with permissions link
		testUUID := uuid.New()
		_, err = s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			testUUID, "Test User", "findbyid@example.com", string(constants.RoleAdmin), now.ToPointer(), now.ToPointer(),
		)
		s.Require().NoError(err)

		// 4. Link user → permissions (admin + user perms)
		_, err = s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2), ($1, $3)`,
			testUUID, adminPermID, userPermID,
		)
		s.Require().NoError(err)

		// 5. Find by ID via repository
		found, err := s.repository.FindByID(s.Ctx, testUUID.String())
		s.Require().NoError(err)
		s.Require().NotNil(found)

		// 6. Assert user core fields
		s.Equal(testUUID, found.ID)
		s.Equal("Test User", found.Name.ValueOrZero())
		s.Equal("findbyid@example.com", found.Email.ValueOrZero())
		s.Equal(string(constants.RoleAdmin), string(found.Role.Data()))

		// 7. Assert permissions loaded from LEFT JOIN + json_agg
		perms := found.Permissions.Data()
		s.Len(perms, 2)
		permMap := make(map[uuid.UUID]*models.Permission, len(perms))
		for _, p := range perms {
			permMap[p.Id] = p
		}

		s.Contains(permMap, adminPermID, "admin permission should be loaded")
		s.Equal(string(constants.RoleAdmin), permMap[adminPermID].Name.ValueOrZero())

		s.Contains(permMap, userPermID, "user permission should be loaded")
		s.Equal(string(constants.RoleUser), permMap[userPermID].Name.ValueOrZero())
	})

	s.Run("success: returns user with no permissions when no links exist", func() {
		// 1. Insert only a user (no permission links)
		testUUID := uuid.New()
		now := utilModels.NewTimestampFromNow()
		_, err := s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			testUUID, "No Perm User", "noperm@example.com", string(constants.RoleUser), now.ToPointer(), now.ToPointer(),
		)
		s.Require().NoError(err)

		// 2. Find by ID — LEFT JOIN with subquery returns empty permissions array
		found, err := s.repository.FindByID(s.Ctx, testUUID.String())
		s.Require().NoError(err)
		s.Require().NotNil(found)

		// 3. Assert user fields exist but permissions is empty slice (not nil)
		s.Equal(testUUID, found.ID)
		perms := found.Permissions.Data()
		s.NotNil(perms, "permissions should not be nil even when no links exist")
		s.Len(perms, 0, "LEFT JOIN with no permission links should return empty array []")
	})

	s.Run("success: returns user when found by email", func() {
		// 1. Insert permission + user
		permID := uuid.MustParse("b0000000-0000-0000-0000-000000000003")
		testUUID := uuid.New()
		now := utilModels.NewTimestampFromNow()

		_, err := s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO permissions (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
			permID, string(constants.RoleModerator), now.ToPointer(), now.ToPointer(),
		)
		s.Require().NoError(err)

		_, err = s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			testUUID, "Email User", "findbyemail@example.com", string(constants.RoleModerator), now.ToPointer(), now.ToPointer(),
		)
		s.Require().NoError(err)

		_, err = s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2)`,
			testUUID, permID,
		)
		s.Require().NoError(err)

		// 2. Find by email
		found, err := s.repository.FindByID(s.Ctx, "findbyemail@example.com")
		s.Require().NoError(err)
		s.Require().NotNil(found)

		// 3. Assert fields
		s.Equal(testUUID, found.ID)
		s.Equal("Email User", found.Name.ValueOrZero())
		s.Len(found.Permissions.Data(), 1)
		s.Equal(string(constants.RoleModerator), found.Permissions.Data()[0].Name.ValueOrZero())
	})

	s.Run("success: returns nil when user not found", func() {
		nonExistentID := uuid.New().String()

		found, err := s.repository.FindByID(s.Ctx, nonExistentID)
		s.Require().NoError(err)
		s.Nil(found)
	})
}

// ==============================================================================
// Save Tests — upserts user + manages permissions in a transaction
// ==============================================================================

func (s *UserRepositorySuite) TestSave_CreateWithPermissions() {
	// 1. Seed permission records
	adminPermID := uuid.MustParse("c0000000-0000-0000-0000-000000000001")
	modPermID := uuid.MustParse("c0000000-0000-0000-0000-000000000002")

	now := utilModels.NewTimestampFromNow()
	_, err := s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO permissions (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		adminPermID, string(constants.RoleAdmin), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	_, err = s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO permissions (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		modPermID, string(constants.RoleModerator), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	// 2. Create user model with permission IDs
	testUUID := uuid.New()
	createdAt := utilModels.NewTimestampFromNow()
	user := &models.User{
		ID:        testUUID,
		Name:      zero.StringFrom("SavePermUser"),
		Email:     zero.StringFrom("saveperm@example.com"),
		Role:      utilModels.NewEnumScan[constants.Role](string(constants.RoleAdmin)),
		CreatedAt: createdAt.ToPointer(),
		UpdatedAt: createdAt.ToPointer(),
		Permissions: utilModels.NewJsonScan[[]*models.Permission]([]*models.Permission{
			{Id: adminPermID, Name: zero.StringFrom(string(constants.RoleAdmin))},
			{Id: modPermID, Name: zero.StringFrom(string(constants.RoleModerator))},
		}),
	}

	// 3. Call Save
	err = s.repository.Save(s.Ctx, user)
	s.Require().NoError(err)

	// 4. Verify user exists via FindByID
	found, err := s.repository.FindByID(s.Ctx, testUUID.String())
	s.Require().NoError(err)
	s.NotNil(found)
	s.Equal(testUUID, found.ID)
	s.Equal("SavePermUser", found.Name.ValueOrZero())

	// 5. Verify permissions were linked (via LEFT JOIN)
	perms := found.Permissions.Data()
	s.Len(perms, 2, "should have 2 permission links after Save")
}

func (s *UserRepositorySuite) TestSave_UpdateExistingUser() {
	// 1. Seed BOTH permission records upfront (FK constraint requires them)
	adminPermID := uuid.MustParse("d0000000-0000-0000-0000-000000000001")
	userPermID := uuid.MustParse("d0000000-0000-0000-0000-000000000002")

	now := utilModels.NewTimestampFromNow()
	_, err := s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO permissions (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		adminPermID, string(constants.RoleAdmin), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	_, err = s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO permissions (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		userPermID, string(constants.RoleUser), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	testUUID := uuid.New()
	createdAt := utilModels.NewTimestampFromNow()

	// Insert initial user with 1 permission link
	_, err = s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		testUUID, "Original", "original@example.com", string(constants.RoleUser), createdAt.ToPointer(), createdAt.ToPointer(),
	)
	s.Require().NoError(err)

	_, err = s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2)`,
		testUUID, adminPermID,
	)
	s.Require().NoError(err)

	// 2. Verify initial state: 1 permission
	found, err := s.repository.FindByID(s.Ctx, testUUID.String())
	s.Require().NoError(err)
	s.Len(found.Permissions.Data(), 1, "should start with 1 permission")

	// 3. Update user with DIFFERENT permissions via Save
	user := &models.User{
		ID:        testUUID,
		Name:      zero.StringFrom("Updated"),
		Email:     zero.StringFrom("updated@example.com"),
		Role:      utilModels.NewEnumScan[constants.Role](string(constants.RoleAdmin)),
		CreatedAt: createdAt.ToPointer(),
		UpdatedAt: utilModels.NewTimestampFromNow().ToPointer(),
		Permissions: utilModels.NewJsonScan[[]*models.Permission]([]*models.Permission{
			{Id: userPermID, Name: zero.StringFrom(string(constants.RoleUser))},
		}),
	}

	err = s.repository.Save(s.Ctx, user)
	s.Require().NoError(err)

	// 4. Verify: name/email updated, OLD permission link removed, NEW permission link added
	found, err = s.repository.FindByID(s.Ctx, testUUID.String())
	s.Require().NoError(err)
	s.Equal("Updated", found.Name.ValueOrZero())
	s.Equal("updated@example.com", found.Email.ValueOrZero())
	s.Len(found.Permissions.Data(), 1, "should have exactly 1 permission after update (old removed, new added)")
	s.Equal(userPermID, found.Permissions.Data()[0].Id,
		"should have the NEW permission ID, not the original")
}

func (s *UserRepositorySuite) TestSave_CreateWithoutPermissions() {
	// 1. Insert user without any permission links
	testUUID := uuid.New()
	now := utilModels.NewTimestampFromNow()
	user := &models.User{
		ID:        testUUID,
		Name:      zero.StringFrom("NoPermUser"),
		Email:     zero.StringFrom("noperm@example.com"),
		Role:      utilModels.NewEnumScan[constants.Role](string(constants.RoleUser)),
		CreatedAt: now.ToPointer(),
		UpdatedAt: now.ToPointer(),
		Permissions: utilModels.NewJsonScan[[]*models.Permission](nil),
	}

	// 2. Call Save
	err := s.repository.Save(s.Ctx, user)
	s.Require().NoError(err)

	// 3. Verify user exists
	found, err := s.repository.FindByID(s.Ctx, testUUID.String())
	s.Require().NoError(err)
	s.NotNil(found)
	s.Equal(testUUID, found.ID)

	// 4. Permissions data: FILTER (WHERE permissions.id IS NOT NULL) returns NULL when no links → COALESCE converts to '[]'
	perms := found.Permissions.Data()
	s.NotNil(perms)
	s.Len(perms, 0, "FILTER + COALESCE should return empty array [] when no permission links")
}

// ==============================================================================
// Delete Tests — soft delete via deleted_at = NOW()
// ==============================================================================

func (s *UserRepositorySuite) TestDelete_ExistingUser() {
	// 1. Insert user directly via SQL (no permission links needed)
	testUUID := uuid.New()
	now := utilModels.NewTimestampFromNow()
	_, err := s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		testUUID, "DeleteUser", "delete@example.com", string(constants.RoleUser), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	// 2. Verify user is visible before delete
	found, err := s.repository.FindByID(s.Ctx, testUUID.String())
	s.Require().NoError(err)
	s.NotNil(found, "user should exist before deletion")

	// 3. Call Delete (soft delete sets deleted_at = NOW())
	err = s.repository.Delete(s.Ctx, testUUID)
	s.Require().NoError(err)

	// 4. Verify user is no longer returned by FindByID (deleted_at IS NULL filter)
	found, err = s.repository.FindByID(s.Ctx, testUUID.String())
	s.Require().NoError(err)
	s.Nil(found, "user should not be found after soft delete")
}

func (s *UserRepositorySuite) TestDelete_NonExistingUser() {
	// 1. Do NOT insert any user — table is clean from SetupTest truncate

	// 2. Call Delete on non-existent UUID
	nonExistent := uuid.New()
	err := s.repository.Delete(s.Ctx, nonExistent)

	// 3. Verify: UPDATE sets affected rows = 0, but the code does NOT check
	//    affected rows — it just commits. So this returns nil (no SQL error).
	// This is the actual behavior of the current implementation.
	s.Require().NoError(err,
		"Delete on non-existent user should not return SQL error (soft delete, no affected-rows check)")
}

func (s *UserRepositorySuite) TestDelete_UserWithPermissions() {
	// 1. Insert user + permissions via SQL
	testUUID := uuid.New()
	adminPermID := uuid.MustParse("e0000000-0000-0000-0000-000000000001")
	now := utilModels.NewTimestampFromNow()

	_, err := s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO permissions (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		adminPermID, string(constants.RoleAdmin), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	_, err = s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		testUUID, "PermUser", "permuser@example.com", string(constants.RoleAdmin), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	_, err = s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2)`,
		testUUID, adminPermID,
	)
	s.Require().NoError(err)

	// 2. Verify user is visible with permissions
	found, err := s.repository.FindByID(s.Ctx, testUUID.String())
	s.Require().NoError(err)
	s.NotNil(found)
	s.Len(found.Permissions.Data(), 1)

	// 3. Delete the user (soft delete only — does NOT clean up user_permissions FK)
	err = s.repository.Delete(s.Ctx, testUUID)
	s.Require().NoError(err)

	// 4. Verify user is no longer returned by FindByID (deleted_at IS NULL)
	found, err = s.repository.FindByID(s.Ctx, testUUID.String())
	s.Require().NoError(err)
	s.Nil(found,
		"user should not be found after soft delete regardless of permission links")
}

func (s *UserRepositorySuite) TestDelete_VerifyDeletedAtIsSet() {
	// 1. Insert user with explicit NULL deleted_at via SQL
	testUUID := uuid.New()
	now := utilModels.NewTimestampFromNow()
	_, err := s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO users (id, name, email, role, created_at, updated_at, deleted_at) VALUES ($1, $2, $3, $4, $5, $6, NULL)`,
		testUUID, "DeletedAtCheck", "deletedat@example.com", string(constants.RoleUser), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	// 2. Verify deleted_at is NULL before delete
	var deletedAtStr *string
	err = s.client.GetClient().QueryRowContext(s.Ctx,
		`SELECT deleted_at::text FROM users WHERE id = $1`, testUUID).Scan(&deletedAtStr)
	s.Require().NoError(err)
	s.Nil(deletedAtStr, "deleted_at should be NULL before delete")

	// 3. Call Delete (soft delete sets deleted_at = NOW())
	err = s.repository.Delete(s.Ctx, testUUID)
	s.Require().NoError(err)

	// 4. Verify deleted_at is now set
	var deletedAtStr2 *string
	err = s.client.GetClient().QueryRowContext(s.Ctx,
		`SELECT deleted_at::text FROM users WHERE id = $1`, testUUID).Scan(&deletedAtStr2)
	s.Require().NoError(err)
	s.NotNil(deletedAtStr2, "deleted_at should be set after soft delete")
}

// ==============================================================================
// List Tests — paginated results + search + role filter + LEFT JOIN permissions
// ==============================================================================

func (s *UserRepositorySuite) TestList_Empty() {
	// 1. Table is clean from SetupTest truncate — no inserts

	// 2. Call List with default params
	params := user.ListParams{Page: 1, PageSize: 10}
	paginator, err := s.repository.List(s.Ctx, params)

	// 3. Verify empty result set (not nil)
	s.Require().NoError(err)
	s.NotNil(paginator.Documents)
	s.Len(paginator.Documents, 0, "empty repository should return empty slice")
	s.Equal(int64(0), paginator.TotalDocument, "total documents should be 0 for empty set")
	s.Equal(int64(0), paginator.TotalPage)
}

func (s *UserRepositorySuite) TestList_SinglePage() {
	// 1. Insert multiple users with permissions directly via SQL
	adminPermID := uuid.MustParse("f0000000-0000-0000-0000-000000000001")
	now := utilModels.NewTimestampFromNow()

	_, err := s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO permissions (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		adminPermID, string(constants.RoleAdmin), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	users := []struct {
		id       uuid.UUID
		name     string
		email    string
		role     constants.Role
	}{
		{uuid.New(), "User A", "userA@example.com", constants.RoleAdmin},
		{uuid.New(), "User B", "userB@example.com", constants.RoleUser},
		{uuid.New(), "User C", "userC@example.com", constants.RoleModerator},
	}

	for _, u := range users {
		_, err = s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			u.id, u.name, u.email, string(u.role), now.ToPointer(), now.ToPointer(),
		)
		s.Require().NoError(err)

		_, err = s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2)`,
			u.id, adminPermID,
		)
		s.Require().NoError(err)
	}

	// 2. Call List — all 3 users fit in single page
	params := user.ListParams{Page: 1, PageSize: 10}
	paginator, err := s.repository.List(s.Ctx, params)

	// 3. Verify results
	s.Require().NoError(err)
	s.NotNil(paginator.Documents)
	s.Len(paginator.Documents, 3, "should return all 3 users")
	// TotalPage should be ceil(3/10) = 1
	s.Equal(int64(1), paginator.TotalPage, "should fit in 1 page")

	// Verify names present (order-independent)
	names := make(map[string]bool)
	for _, u := range paginator.Documents {
		names[u.Name.ValueOrZero()] = true
	}
	s.True(names["User A"])
	s.True(names["User B"])
	s.True(names["User C"])
}

func (s *UserRepositorySuite) TestList_Pagination() {
	// 1. Insert 5 users directly via SQL
	now := utilModels.NewTimestampFromNow()
	var ids []uuid.UUID

	for i := int64(0); i < 5; i++ {
		id := uuid.New()
		ids = append(ids, id)
		_, err := s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			id, fmt.Sprintf("PUser %d", i), fmt.Sprintf("puser%d@example.com", i), string(constants.RoleUser), now.ToPointer(), now.ToPointer(),
		)
		s.Require().NoError(err)
	}

	// 2. Page 1 with pageSize=2
	params := user.ListParams{Page: 1, PageSize: 2}
	paginator, err := s.repository.List(s.Ctx, params)
	s.Require().NoError(err)

	s.Len(paginator.Documents, 2, "page 1 should return 2 users")
	s.Equal(int64(5), paginator.TotalDocument, "total should still be 5")
	s.Equal(int64(3), paginator.TotalPage, "5 items / 2 per page = 3 pages (ceiling)")

	// DB order is not deterministic without ORDER BY, so we verify all expected IDs
	// appear exactly once across all pages.

	// 3. Page 2 with pageSize=2
	params.Page = 2
	paginator, err = s.repository.List(s.Ctx, params)
	s.Require().NoError(err)

	s.Len(paginator.Documents, 2, "page 2 should return 2 users")

	// 4. Page 3 (last page with remaining items)
	params.Page = 3
	paginator, err = s.repository.List(s.Ctx, params)
	s.Require().NoError(err)

	s.Len(paginator.Documents, 1, "page 3 should return last user")

	// Verify: across all 3 pages we get exactly the 5 inserted IDs,
	// no duplicates, each appears once.
	seen := make(map[uuid.UUID]int) // uuid -> page number (1-indexed)

	for _, u := range paginator.Documents {
		if pg, ok := seen[u.ID]; ok {
			s.Failf("duplicate ID", "UUID %s found on both page 3 and page %d", u.ID, pg)
		}
		seen[u.ID] = 3
	}

	// Re-run pages 1 and 2 to collect their IDs into the map.
	params.Page = 1
	paginator, err = s.repository.List(s.Ctx, params)
	s.Require().NoError(err)
	for _, u := range paginator.Documents {
		if pg, ok := seen[u.ID]; ok {
			s.Failf("duplicate ID", "UUID %s found on both page 1 and page %d", u.ID, pg)
		}
		seen[u.ID] = 1
	}

	params.Page = 2
	paginator, err = s.repository.List(s.Ctx, params)
	s.Require().NoError(err)
	for _, u := range paginator.Documents {
		if pg, ok := seen[u.ID]; ok {
			s.Failf("duplicate ID", "UUID %s found on both page 2 and page %d", u.ID, pg)
		}
		seen[u.ID] = 2
	}

	s.Len(seen, 5, "all 5 expected IDs should appear exactly once across all pages")
	for _, id := range ids {
		_, ok := seen[id]
		s.True(ok, "UUID %s should be present on some page", id)
	}
}

func (s *UserRepositorySuite) TestList_SearchwordFilter() {
	// 1. Insert users with different email domains via SQL
	now := utilModels.NewTimestampFromNow()

	_, err := s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), "Alpha", "alpha@example.com", string(constants.RoleAdmin), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	_, err = s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), "Beta", "beta@test.com", string(constants.RoleUser), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	_, err = s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
		uuid.New(), "Gamma", "gamma@example.com", string(constants.RoleModerator), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	// 2. Search by email containing "example"
	params := user.ListParams{Page: 1, PageSize: 10, Searchword: "example"}
	paginator, err := s.repository.List(s.Ctx, params)

	s.Require().NoError(err)
	s.Len(paginator.Documents, 2,
		"should return 2 users with 'example' in email (Alpha + Gamma)")

	// Verify both are from example.com domain
	for _, u := range paginator.Documents {
		s.Contains(u.Email.ValueOrZero(), "example",
			"each result should contain 'example' in email")
	}
}

func (s *UserRepositorySuite) TestList_RoleFilter() {
	// 1. Insert users with different roles via SQL
	adminPermID := uuid.MustParse("a0000000-0000-0000-0000-000000000099")
	now := utilModels.NewTimestampFromNow()

	_, err := s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO permissions (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		adminPermID, string(constants.RoleAdmin), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	// Insert 2 Admin users and 2 User roles
	for i := int64(0); i < 2; i++ {
		id := uuid.New()
		_, err = s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			id, fmt.Sprintf("Admin %d", i), fmt.Sprintf("admin%d@example.com", i), string(constants.RoleAdmin), now.ToPointer(), now.ToPointer(),
		)
		s.Require().NoError(err)

		_, err = s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2)`,
			id, adminPermID,
		)
		s.Require().NoError(err)
	}

	// 3. Insert User role users — different permission ID (not "admin" perm)
	userPermID := uuid.MustParse("b0000000-0000-0000-0000-000000000098")
	_, err = s.client.GetClient().ExecContext(s.Ctx,
		`INSERT INTO permissions (id, name, created_at, updated_at) VALUES ($1, $2, $3, $4)`,
		userPermID, string(constants.RoleUser), now.ToPointer(), now.ToPointer(),
	)
	s.Require().NoError(err)

	for i := int64(0); i < 2; i++ {
		id := uuid.New()
		_, err = s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO users (id, name, email, role, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`,
			id, fmt.Sprintf("User %d", i), fmt.Sprintf("user%d@example.com", i), string(constants.RoleUser), now.ToPointer(), now.ToPointer(),
		)
		s.Require().NoError(err)

		_, err = s.client.GetClient().ExecContext(s.Ctx,
			`INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2)`,
			id, userPermID,
		)
		s.Require().NoError(err)
	}

	// 4. Filter by Role = Admin
	params := user.ListParams{Page: 1, PageSize: 10, Role: constants.RoleAdmin}
	paginator, err := s.repository.List(s.Ctx, params)

	s.Require().NoError(err)
	s.Len(paginator.Documents, 2,
		"should return only 2 Admin users when filtered by Role")

	for _, u := range paginator.Documents {
		s.Equal(string(constants.RoleAdmin), string(u.Role.Data()),
			"all results should have Admin role")
	}
}
