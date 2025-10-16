package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	mockdb "github.com/joekings2k/logistics-eta/db/mock"
	db "github.com/joekings2k/logistics-eta/db/sqlc"
	"github.com/joekings2k/logistics-eta/util"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)
func randomUser(t *testing.T) (user db.User, password string) {
	password= util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)
	user = db.User{
		ID: uuid.New(),
		Email:util.RandomEmail(),
		Name: util.RandomString(10),
		PasswordHash: hashedPassword,
		Role: util.RandomRole(),
	}
	return user, password
}

type eqCreateUserParamsMatcher struct {
	arg db.CreateUserParams
	password string
}

func (e eqCreateUserParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateUserParams)
	if !ok {
		return false
	}
	if arg.ID == uuid.Nil {
	return false
}
	err := util.CheckPassword(e.password, arg.PasswordHash )
	if err != nil {
		return false
	}
	e.arg.ID = arg.ID
	e.arg.PasswordHash = arg.PasswordHash
	return  reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateUserParamsMatcher) String()string {
	return fmt.Sprintf("matches arg  %v and password %v", e.arg ,e.password)
}

func EqCreateUserParams (arg db.CreateUserParams, password string ) gomock.Matcher {
	return eqCreateUserParamsMatcher{arg,password}
}
func TestCreateUser(t *testing.T) {
	user, password := randomUser(t)
	testCases := []struct{
		name string
		body gin.H
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(t *testing.T ,recorder *httptest.ResponseRecorder)
	}{
		{
			name : "OK",
			body: gin.H{
				"name": user.Name,
				"email": user.Email,
				"password": password,
				"role": user.Role,
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateUserParams{
					ID: user.ID,
					Name: user.Name,
					Email: user.Email,
					PasswordHash: user.PasswordHash,
					Role: user.Role,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchUser(t, recorder.Body, user)
			},
		},
		{
			name : "InternalServerError",
			body: gin.H{
				"name": user.Name,
				"email": user.Email,
				"password": password,
				"role": user.Role,
			},
			buildStubs: func(store *mockdb.MockStore) {
				
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name : "InvalidID",
			body: gin.H{
				"name": user.Name,
				"email": user.Email,
				"password": password,
				"role": user.Role,
			},
			buildStubs: func(store *mockdb.MockStore) {		
				arg := db.CreateUserParams{
					Name: user.Name,
					Email: user.Email,
					PasswordHash: user.PasswordHash,
					Role: user.Role,
				}
				store.EXPECT().
					CreateUser(gomock.Any(), EqCreateUserParams(arg, password)).
					Times(1).
					Return(db.User{}, sql.ErrTxDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "InvalidPrams",
			body: gin.H{
				"name": "",
				"email": "",
				"password": "",
				"role": "",
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
			},
		},
		{
			name: "DuplicateEmail",
			body: gin.H{
				"name": user.Name,
				"email": user.Email,
				"password": password,
				"role": user.Role,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					CreateUser(gomock.Any(),gomock.Any()).Times(1).Return(db.User{}, &pq.Error{Code: "23505"})
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusForbidden, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store:= mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()
			url := "/users/register"
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)

		})
	}
}

func TestLoginUser(t *testing.T) {
	user, password := randomUser(t)
	user.Role = string(util.RoleCustomer)

	testCases := []struct {
		name string
		body gin.H
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
    {
			name: "OK",
			body: gin.H{
				"email":user.Email,
				"password": password,
				"role": user.Role,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(user, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
    {
			name: "UserNotFound",
			body: gin.H{
				"email":"unknownUser@gmail.com",
				"password": password,
				"role": user.Role,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
			},
		},
    {
			name: "IncorrectPassword",
			body: gin.H{
				"email":user.Email,
				"password":"wrongpassword",
				"role": user.Role,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db.User{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"email":    user.Email,
				"password": password,
				"role":     user.Role,
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.User{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusInternalServerError, recorder.Code)
			},
		},
		{
			name: "RoleMismatch",
			body: gin.H{
				"email":    user.Email,
				"password": password,
				"role":     string(util.RoleAdmin),
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetUserByEmail(gomock.Any(), gomock.Eq(user.Email)).
					Times(1).
					Return(db.User{}, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
			require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
	}
	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			store:= mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()
			url := "/users/login"
			data, err := json.Marshal(tc.body)
			require.NoError(t, err)
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)

		})
	}
}


func requireBodyMatchUser(t *testing.T, body *bytes.Buffer, user db.User) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)
	var gotUser UserResponse
	err = json.Unmarshal(data, &gotUser)
	require.NoError(t, err)
	
	require.Equal(t, user.Email, gotUser.Email)
	require.Equal(t, user.Name, gotUser.Name)
	require.Equal(t, user.Role, gotUser.Role)

}