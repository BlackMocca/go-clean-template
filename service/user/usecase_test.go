package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"example/constants"
	"example/models"
	"example/service/user"
	mocks "example/service/user/mocks"

	utilModels "github.com/BlackMocca/utils/models"
	"github.com/google/uuid"
	"github.com/guregu/null/v6/zero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------
func zeroString(s string) zero.String { return zero.StringFrom(s) }

func buildUser(id uuid.UUID, email string, role constants.Role) *models.User {
	now := utilModels.NewTimestampFromNow()
	return &models.User{
		ID:        id,
		Name:      zeroString("John Doe"),
		Email:     zeroString(email),
		Role:      utilModels.NewEnumScan[constants.Role](string(role)),
		CreatedAt: now.ToPointer(),
		UpdatedAt: now.ToPointer(),
	}
}

// assertFullUser asserts every field on models.User individually.
func assertFullUser(t *testing.T, expected, actual *models.User) {
	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Email, actual.Email)
	assert.Equal(t, expected.Role.Data(), actual.Role.Data())

	// Timestamp fields — compare via time.Time
	if expected.CreatedAt != nil && actual.CreatedAt != nil {
		assert.True(t, time.Time(*expected.CreatedAt).Equal(time.Time(*actual.CreatedAt)))
	} else if expected.CreatedAt == nil && actual.CreatedAt == nil {
		assert.Nil(t, actual.CreatedAt)
	}

	if expected.UpdatedAt != nil && actual.UpdatedAt != nil {
		assert.True(t, time.Time(*expected.UpdatedAt).Equal(time.Time(*actual.UpdatedAt)))
	} else if expected.UpdatedAt == nil && actual.UpdatedAt == nil {
		assert.Nil(t, actual.UpdatedAt)
	}

	assert.Equal(t, expected.DeletedAt, actual.DeletedAt)

	// Permissions slice length + each item
	expPerms := expected.Permissions.Data()
	actPerms := actual.Permissions.Data()
	if expPerms == nil && actPerms == nil {
		assert.Nil(t, actual.Permissions.Data())
	} else {
		require.NotNil(t, actPerms)
		assert.Len(t, expPerms, len(actPerms))
		for i := range expPerms {
			assert.Equal(t, expPerms[i].Id, actPerms[i].Id)
			assert.Equal(t, expPerms[i].Name.ValueOrZero(), actPerms[i].Name.ValueOrZero())
		}
	}
}

// assertFullPaginator asserts every field on Paginator individually.
func assertFullPaginator(t *testing.T, expected, actual models.Paginator[[]*models.User]) {
	assert.Equal(t, expected.Page, actual.Page)
	assert.Equal(t, expected.PageSize, actual.PageSize)
	assert.Equal(t, expected.TotalPage, actual.TotalPage)
	assert.Equal(t, expected.TotalDocument, actual.TotalDocument)

	expDocs := expected.Documents
	actDocs := actual.Documents
	require.NotNil(t, actDocs)
	assert.Len(t, expDocs, len(actDocs))
	for i := range expDocs {
		assertFullUser(t, expDocs[i], actDocs[i])
	}
}

// ===========================================================================
// GetUserByID
// ===========================================================================
func TestUserUsecase_GetUserByID(t *testing.T) {
	t.Run("success: retrieves user by UUID", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		testID := uuid.MustParse("11111111-2222-3333-4444-555555555555")
		expected := buildUser(testID, "john@example.com", constants.RoleAdmin)

		mockRepo.EXPECT().
			FindByID(mock.Anything, testID.String()).
			Run(func(ctx context.Context, idOrEmail string) {
				assert.NotNil(t, ctx)
				assert.Equal(t, "11111111-2222-3333-4444-555555555555", idOrEmail)
			}).
			Return(expected, nil).Times(1)

		result, err := uc.GetUserByID(context.Background(), testID)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assertFullUser(t, expected, result)
	})

	t.Run("success: returns nil when user not found (nil user, no error)", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		testID := uuid.MustParse("22222222-3333-4444-5555-666666666666")

		mockRepo.EXPECT().
			FindByID(mock.Anything, testID.String()).
			Run(func(ctx context.Context, idOrEmail string) {
				assert.NotNil(t, ctx)
				assert.Equal(t, "22222222-3333-4444-5555-666666666666", idOrEmail)
			}).
			Return(nil, nil).Times(1)

		result, err := uc.GetUserByID(context.Background(), testID)

		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("failure: propagates repository error", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		testID := uuid.MustParse("33333333-4444-5555-6666-777777777777")
		expectedErr := errors.New("connection timeout: dial tcp 10.0.0.1:5432: i/o timeout")

		mockRepo.EXPECT().
			FindByID(mock.Anything, testID.String()).
			Run(func(ctx context.Context, idOrEmail string) {
				assert.NotNil(t, ctx)
				assert.Equal(t, "33333333-4444-5555-6666-777777777777", idOrEmail)
			}).
			Return(nil, expectedErr).Times(1)

		result, err := uc.GetUserByID(context.Background(), testID)

		require.Error(t, err)
		assert.Equal(t, "connection timeout: dial tcp 10.0.0.1:5432: i/o timeout", err.Error())
		assert.Nil(t, result)
		mockRepo.AssertNumberOfCalls(t, "FindByID", 1)
	})
}

// ===========================================================================
// ListUsers
// ===========================================================================
func TestUserUsecase_ListUsers(t *testing.T) {
	t.Run("success: returns paginated users with correct params", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		testUsers := []*models.User{
			buildUser(uuid.MustParse("aaaa1111-2222-3333-4444-555555555555"), "alice@example.com", constants.RoleAdmin),
			buildUser(uuid.MustParse("bbbb2222-3333-4444-5555-666666666666"), "bob@example.com", constants.RoleUser),
		}

		expectedPaginator := models.Paginator[[]*models.User]{
			Page:          1,
			PageSize:      20,
			TotalDocument: 2,
			TotalPage:     1,
			Documents:     testUsers,
		}

		expectedParams := user.ListParams{
			Page:       1,
			PageSize:   20,
			Searchword: "ali",
			Role:       constants.RoleAdmin,
		}

		mockRepo.EXPECT().
			List(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, params user.ListParams) {
				assert.NotNil(t, ctx)
				assert.Equal(t, int64(1), params.Page)
				assert.Equal(t, int64(20), params.PageSize)
				assert.Equal(t, "ali", params.Searchword)
				assert.Equal(t, constants.RoleAdmin, params.Role)
			}).
			Return(expectedPaginator, nil).Times(1)

		result, err := uc.ListUsers(context.Background(), expectedParams)

		require.NoError(t, err)
		require.NotNil(t, result)
		assertFullPaginator(t, expectedPaginator, result)
	})

	t.Run("success: empty documents slice (not nil)", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		expectedParams := user.ListParams{Page: 1, PageSize: 10}

		mockRepo.EXPECT().
			List(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, params user.ListParams) {
				assert.NotNil(t, ctx)
				assert.Equal(t, int64(1), params.Page)
				assert.Equal(t, int64(10), params.PageSize)
				assert.Empty(t, params.Searchword)
				var emptyRole constants.Role
				assert.Equal(t, emptyRole, params.Role)
			}).
			Return(models.Paginator[[]*models.User]{
				Page:          1,
				PageSize:      10,
				TotalDocument: 0,
				TotalPage:     0,
				Documents:     []*models.User{},
			}, nil).Times(1)

		result, err := uc.ListUsers(context.Background(), expectedParams)

		require.NoError(t, err)
		assert.NotNil(t, result.Documents)
		assert.Len(t, result.Documents, 0)
		assert.Equal(t, int64(0), result.TotalDocument)
		assert.Equal(t, int64(0), result.TotalPage)
	})

	t.Run("failure: propagates repository error", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		expectedErr := errors.New("query timeout after 30s")
		params := user.ListParams{Page: 1, PageSize: 5}

		mockRepo.EXPECT().
			List(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, p user.ListParams) {
				assert.NotNil(t, ctx)
				assert.Equal(t, int64(1), p.Page)
				assert.Equal(t, int64(5), p.PageSize)
			}).
			Return(models.Paginator[[]*models.User]{}, expectedErr).Times(1)

		result, err := uc.ListUsers(context.Background(), params)

		require.Error(t, err)
		assert.Equal(t, "query timeout after 30s", err.Error())
		var zeroPaginator models.Paginator[[]*models.User]
		assert.Equal(t, zeroPaginator, result)
		mockRepo.AssertNumberOfCalls(t, "List", 1)
	})

	t.Run("success: different pagination values passed correctly", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		testUsers := []*models.User{buildUser(uuid.New(), "x@example.com", constants.RoleUser)}
		expectedPaginator := models.Paginator[[]*models.User]{
			Page:          3, PageSize: 50, TotalDocument: 127, TotalPage: 3, Documents: testUsers,
		}

		inputParams := user.ListParams{Page: 3, PageSize: 50, Searchword: "test", Role: constants.RoleUser}

		mockRepo.EXPECT().
			List(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, params user.ListParams) {
				assert.Equal(t, int64(3), params.Page)
				assert.Equal(t, int64(50), params.PageSize)
				assert.Equal(t, "test", params.Searchword)
				assert.Equal(t, constants.RoleUser, params.Role)
			}).
			Return(expectedPaginator, nil).Times(1)

		result, err := uc.ListUsers(context.Background(), inputParams)

		require.NoError(t, err)
		assertFullPaginator(t, expectedPaginator, result)
	})
}

// ===========================================================================
// Save
// ===========================================================================
func TestUserUsecase_Save(t *testing.T) {
	t.Run("success: saves user when email is unique", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		newUser := buildUser(uuid.MustParse("cccc3333-4444-5555-6666-777777777777"), "newuser@example.com", constants.RoleUser)

		// Call 1: FindByID for email uniqueness — returns nil (unique)
		mockRepo.EXPECT().
			FindByID(mock.Anything, newUser.Email.ValueOrZero()).
			Run(func(ctx context.Context, idOrEmail string) {
				assert.NotNil(t, ctx)
				assert.Equal(t, "newuser@example.com", idOrEmail)
			}).
			Return(nil, nil).Times(1)

		// Call 2: Save the user
		mockRepo.EXPECT().
			Save(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, u *models.User) {
				assert.NotNil(t, ctx)
				assert.Equal(t, newUser.ID, u.ID)
				assert.Equal(t, zeroString("John Doe"), u.Name)
				assert.Equal(t, zeroString("newuser@example.com"), u.Email)
				assert.Equal(t, constants.Role(string(constants.RoleUser)), u.Role.Data())
				require.NotNil(t, u.CreatedAt)
				require.NotNil(t, u.UpdatedAt)
			}).
			Return(nil).Times(1)

		err := uc.Save(context.Background(), newUser)

		require.NoError(t, err)
		mockRepo.AssertNumberOfCalls(t, "FindByID", 1)
		mockRepo.AssertNumberOfCalls(t, "Save", 1)
	})

	t.Run("failure: returns conflict error when email is duplicated", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		newUser := buildUser(uuid.MustParse("dddd4444-5555-6666-7777-888888888888"), "taken@example.com", constants.RoleAdmin)
		existingUser := buildUser(uuid.MustParse("eeee5555-6666-7777-8888-999999999999"), "taken@example.com", constants.RoleAdmin)

		mockRepo.EXPECT().
			FindByID(mock.Anything, newUser.Email.ValueOrZero()).
			Run(func(ctx context.Context, idOrEmail string) {
				assert.NotNil(t, ctx)
				assert.Equal(t, "taken@example.com", idOrEmail)
			}).
			Return(existingUser, nil).Times(1)

		err := uc.Save(context.Background(), newUser)

		require.Error(t, err)
		assert.Equal(t, string(constants.ApplicationCodeConflictUserEmailWasDuplicate), err.Error())
		mockRepo.AssertNumberOfCalls(t, "FindByID", 1)
		mockRepo.AssertNumberOfCalls(t, "Save", 0)
	})

	t.Run("failure: propagates FindByID error (does not call Save)", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		newUser := buildUser(uuid.MustParse("ffff6666-7777-8888-9999-aaaaaaaaaaaa"), "fail@example.com", constants.RoleUser)
		dbErr := errors.New("connection pool exhausted: max connections reached (100)")

		mockRepo.EXPECT().
			FindByID(mock.Anything, newUser.Email.ValueOrZero()).
			Run(func(ctx context.Context, idOrEmail string) {
				assert.NotNil(t, ctx)
				assert.Equal(t, "fail@example.com", idOrEmail)
			}).
			Return(nil, dbErr).Times(1)

		err := uc.Save(context.Background(), newUser)

		require.Error(t, err)
		assert.Equal(t, "connection pool exhausted: max connections reached (100)", err.Error())
		mockRepo.AssertNumberOfCalls(t, "FindByID", 1)
		mockRepo.AssertNumberOfCalls(t, "Save", 0)
	})
}

// ===========================================================================
// Delete
// ===========================================================================
func TestUserUsecase_Delete(t *testing.T) {
	t.Run("success: deletes existing user", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		testID := uuid.MustParse("1111aaaa-2222-3333-4444-555555555555")
		existingUser := buildUser(testID, "delete@example.com", constants.RoleUser)

		mockRepo.EXPECT().
			FindByID(mock.Anything, testID.String()).
			Run(func(ctx context.Context, idOrEmail string) {
				assert.NotNil(t, ctx)
				assert.Equal(t, "1111aaaa-2222-3333-4444-555555555555", idOrEmail)
			}).
			Return(existingUser, nil).Times(1)

		mockRepo.EXPECT().
			Delete(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, id uuid.UUID) {
				assert.NotNil(t, ctx)
				assert.Equal(t, testID, id)
			}).
			Return(nil).Times(1)

		err := uc.Delete(context.Background(), testID)

		require.NoError(t, err)
		mockRepo.AssertNumberOfCalls(t, "FindByID", 1)
		mockRepo.AssertNumberOfCalls(t, "Delete", 1)
	})

	t.Run("failure: returns conflict when user already removed (FindByID returns nil)", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		testID := uuid.MustParse("2222bbbb-3333-4444-5555-666666666666")

		mockRepo.EXPECT().
			FindByID(mock.Anything, testID.String()).
			Run(func(ctx context.Context, idOrEmail string) {
				assert.NotNil(t, ctx)
				assert.Equal(t, "2222bbbb-3333-4444-5555-666666666666", idOrEmail)
			}).
			Return(nil, nil).Times(1)

		err := uc.Delete(context.Background(), testID)

		require.Error(t, err)
		assert.Equal(t, string(constants.ApplicationCodeConflictUserAlreadyRemoved), err.Error())
		mockRepo.AssertNumberOfCalls(t, "FindByID", 1)
		mockRepo.AssertNumberOfCalls(t, "Delete", 0)
	})

	t.Run("failure: propagates FindByID error (does not call Delete)", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		testID := uuid.MustParse("3333cccc-4444-5555-6666-777777777777")
		dbErr := errors.New("deadlock detected: transaction wait timeout")

		mockRepo.EXPECT().
			FindByID(mock.Anything, testID.String()).
			Run(func(ctx context.Context, idOrEmail string) {
				assert.NotNil(t, ctx)
				assert.Equal(t, "3333cccc-4444-5555-6666-777777777777", idOrEmail)
			}).
			Return(nil, dbErr).Times(1)

		err := uc.Delete(context.Background(), testID)

		require.Error(t, err)
		assert.Equal(t, "deadlock detected: transaction wait timeout", err.Error())
		mockRepo.AssertNumberOfCalls(t, "FindByID", 1)
		mockRepo.AssertNumberOfCalls(t, "Delete", 0)
	})

	t.Run("failure: propagates Delete error after FindByID succeeds", func(t *testing.T) {
		mockRepo := mocks.NewMockRepository(t)
		uc := user.NewUserUsecase(mockRepo)

		testID := uuid.MustParse("4444dddd-5555-6666-7777-888888888888")
		existingUser := buildUser(testID, "delete2@example.com", constants.RoleAdmin)
		deleteErr := errors.New("foreign key constraint violation: permission reference exists")

		mockRepo.EXPECT().
			FindByID(mock.Anything, testID.String()).
			Run(func(ctx context.Context, idOrEmail string) {
				assert.NotNil(t, ctx)
				assert.Equal(t, "4444dddd-5555-6666-7777-888888888888", idOrEmail)
			}).
			Return(existingUser, nil).Times(1)

		mockRepo.EXPECT().
			Delete(mock.Anything, mock.Anything).
			Run(func(ctx context.Context, id uuid.UUID) {
				assert.NotNil(t, ctx)
				assert.Equal(t, testID, id)
			}).
			Return(deleteErr).Times(1)

		err := uc.Delete(context.Background(), testID)

		require.Error(t, err)
		assert.Equal(t, "foreign key constraint violation: permission reference exists", err.Error())
		mockRepo.AssertNumberOfCalls(t, "FindByID", 1)
		mockRepo.AssertNumberOfCalls(t, "Delete", 1)
	})
}
