package db

import (
	"context"
	"database/sql"
	"testing"
	"github.com/google/uuid"
	"github.com/joekings2k/logistics-eta/util"
	"github.com/stretchr/testify/require"
)


func createRandomUser(t *testing.T) User {
	hashedPassword, err := util.HashPassword(util.RandomString(6))
	require.NoError(t, err)
	arg:= CreateUserParams{
		ID: uuid.New(),
		Name: util.RandomString(6),
		Email: util.RandomEmail(),
		PasswordHash:hashedPassword,
		Role: util.RandomRole(),
	}
	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)
	
	require.Equal(t, arg.ID, user.ID)
	require.Equal(t, arg.Name, user.Name)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.PasswordHash, user.PasswordHash)
	require.Equal(t, arg.Role, user.Role)
	return user
}

func TestCreateUser(t *testing.T) {
	createRandomUser(t)
}

func TestGetUserByID(t *testing.T) {
	user1 := createRandomUser(t)	
	user2, err := testQueries.GetUserByID(context.Background(), user1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	
	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Name, user2.Name)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.PasswordHash, user2.PasswordHash)
	require.Equal(t, user1.Role, user2.Role)
}

func TestGetUserByEmail(t *testing.T) {
	user1 := createRandomUser(t)
	user2, err := testQueries.GetUserByEmail(context.Background(), user1.Email)
	require.NoError(t, err )
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Name, user2.Name)
	require.Equal(t, user1.Email, user2.Email)
	require.Equal(t, user1.PasswordHash, user2.PasswordHash)
	require.Equal(t, user1.Role, user2.Role)
	
}


func TestDeleteUser(t *testing.T) {
	user1 := createRandomUser(t)
	err := testQueries.DeleteUser(context.Background(), user1.ID)
	require.NoError(t, err)
	user2, err := testQueries.GetUserByID(context.Background(), user1.ID)
	require.Error(t, err)
	require.Error(t, err, sql.ErrNoRows.Error())
	require.Empty(t, user2)

}

func TestUpdateUser(t *testing.T) {
	user1 := createRandomUser(t)
	newEmail := util.RandomEmail()
	newHashedPassword, err:= util.HashPassword(util.RandomString(6))
	require.NoError(t, err)
	arg:= UpdateUserParams{
		ID: user1.ID,
		Name: "userNewName",
		Email: newEmail,
		PasswordHash: newHashedPassword,
		Role: util.RandomRole(),

	}
	user2, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.Equal(t, user1.ID, user2.ID)

}
func TestUpdateUserInvalidID(t *testing.T) {
	user1 := createRandomUser(t)
	arg:= UpdateUserParams{
		ID: uuid.New(),
		Name: "userNewName",
		Email: user1.Email,
		PasswordHash: user1.PasswordHash,
		Role: user1.Role,

	}
	user2, err := testQueries.UpdateUser(context.Background(), arg)
	require.Error(t, err)
	require.Empty(t, user2)
	require.EqualError(t, err, sql.ErrNoRows.Error())

}

func TestUpdatePartial(t *testing.T) {
	user1 := createRandomUser(t)
	newEmail := util.RandomEmail()

	arg := UpdateUserPartialParams{
		ID: user1.ID,
		Email: sql.NullString{String: newEmail, Valid: true},
	}
	user2, err := testQueries.UpdateUserPartial(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user2)
	require.NotEqual(t, user1.Email, user2.Email)
	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, user1.Name, user2.Name)
	require.Equal(t, user1.Role, user2.Role)
	require.Equal(t, user1.PasswordHash, user2.PasswordHash)
}

func TestListUsers(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomUser(t)
	}
	arg := ListUsersParams{
		Limit: 5,
		Offset: 5,
	}
	users, err := testQueries.ListUsers(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, users)
	require.Len(t, users, 5)

	for _, user := range users {
		require.NotEmpty(t, user)
	}
	arg = ListUsersParams{
		Limit: 0,
		Offset: 0}
	users, err = testQueries.ListUsers(context.Background(), arg)
	require.NoError(t, err)
	require.Len(t, users, 0)
}


// var validArg = UpdateUserParams{
// 	ID: uuid.New(),
// 	Name: "validName",
// 	Email: util.RandomEmail(),
// 	PasswordHash: "validPasswordHash",
// 	Role: "user",
// }

// var invalidIDArg = UpdateUserParams{
// 	ID: uuid.New(),
// 	Name: "validName",
// 	Email: util.RandomEmail(),
// 	PasswordHash: "validPasswordHash",
// 	Role: "user",
// }

// var invalidEmailArg = UpdateUserParams{
// 	ID: uuid.New(),
// 	Name: "validName",
// 	Email: "invalidEmail",
// 	PasswordHash: "validPasswordHash",
// 	Role: "user",
// }

// func TestUpdateUserCases(t *testing.T) {
// 	cases := []struct {
// 		name    string
// 		arg     UpdateUserParams
// 		wantErr bool
// 	}{
// 		{"valid update", validArg, false},
// 		{"invalid user id", invalidIDArg, true},
// 		{"invalid email", invalidEmailArg, true},
// 	}

// 	for _, tc := range cases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			user, err := testQueries.UpdateUser(context.Background(), tc.arg)
// 			if tc.wantErr {
// 				require.Error(t, err)
// 			} else {
// 				require.NoError(t, err)
// 				require.NotEmpty(t, user)
// 			}
// 		})
// 	}
// }