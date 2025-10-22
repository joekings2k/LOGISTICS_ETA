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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	mockdb "github.com/joekings2k/logistics-eta/db/mock"
	db "github.com/joekings2k/logistics-eta/db/sqlc"
	"github.com/joekings2k/logistics-eta/token"
	"github.com/joekings2k/logistics-eta/util"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)


func RandomVehicle(t *testing.T)db.Vehicle{
	user, _ := randomUser(t) 
	return db.Vehicle{
		ID: uuid.New(),
		DriverID: user.ID,
		LicensePlate: util.RandomString(10),
		Model: sql.NullString{String: util.RandomString(6), Valid: true},
		ImageUrl: sql.NullString{String: util.RandomString(6), Valid: true},
		Capacity: sql.NullInt32{Int32: int32(util.RandomInt(1,100)), Valid: true},
	}

}

type eqCreateVehicleParamsMatcher struct{
	arg db.CreateVehicleParams
}

func (e eqCreateVehicleParamsMatcher) Matches(x interface{}) bool {
	arg, ok := x.(db.CreateVehicleParams)
	if !ok{
		return false
	}
	e.arg.ID = arg.ID
	return reflect.DeepEqual(e.arg, arg)
}

func (e eqCreateVehicleParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v", e.arg)
}

func EqCreateVehicleParams(arg db.CreateVehicleParams) gomock.Matcher {
	return eqCreateVehicleParamsMatcher{arg}
}

func TestCreateVehicle(t *testing.T) {
	user, _ := randomUser(t)
	vehicle := RandomVehicle(t)
	vehicle.DriverID = user.ID
	user.Role = string(util.RoleDriver)

	testCases := []struct{
		name string
		body gin.H
		setupAuth func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(t *testing.T ,recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			body: gin.H{
				"license_plate": vehicle.LicensePlate,
				"model": vehicle.Model.String,
				"image_url": vehicle.ImageUrl.String,
				"capacity": vehicle.Capacity.Int32,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateVehicleParams{
					ID: vehicle.ID,
					DriverID: vehicle.DriverID,
					LicensePlate: vehicle.LicensePlate,
					Model: sql.NullString{String: vehicle.Model.String, Valid: true},
					ImageUrl: sql.NullString{String: vehicle.ImageUrl.String, Valid: true},
					Capacity: sql.NullInt32{Int32: vehicle.Capacity.Int32, Valid: true},
				}
				store.EXPECT().GetUserByID(gomock.Any(), gomock.Eq(user.ID)).Times(1).Return(user, nil)
				store.EXPECT().
					CreateVehicle(gomock.Any(), EqCreateVehicleParams(arg)).
					Times(1).
					Return(vehicle, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchVehicle(t, recorder.Body, vehicle)
			},
		},
		{
			name: "InternalErrorUser",
			body: gin.H{
				"license_plate": vehicle.LicensePlate,
				"model": vehicle.Model.String,
				"image_url": vehicle.ImageUrl.String,
				"capacity": vehicle.Capacity.Int32,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetUserByID(gomock.Any(), gomock.Eq(user.ID)).Times(1).Return(db.User{}, sql.ErrConnDone)
				
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				
			},
		},
		{
			name: "InternalError",
			body: gin.H{
				"license_plate": vehicle.LicensePlate,
				"model": vehicle.Model.String,
				"image_url": vehicle.ImageUrl.String,
				"capacity": vehicle.Capacity.Int32,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetUserByID(gomock.Any(), gomock.Eq(user.ID)).Times(1).Return(user, nil)
				store.EXPECT().
					CreateVehicle(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Vehicle{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				
			},
		},
		{
			name: "InvalidPrams",
			body: gin.H{
				"license_plate": "",
				"model": "",
				"image_url": "",
				"capacity": "",
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().GetUserByID(gomock.Any(), gomock.Any()).Times(0)
				store.EXPECT().
					CreateVehicle(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				
			},
		},
		{
			name: "InvalidRole",
			body: gin.H{
				"license_plate": vehicle.LicensePlate,
				"model": vehicle.Model.String,
				"image_url": vehicle.ImageUrl.String,
				"capacity": vehicle.Capacity.Int32,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				
				user.Role = string(util.RoleAdmin)
				store.EXPECT().GetUserByID(gomock.Any(), gomock.Eq(user.ID)).Times(1).Return(user, nil)
				store.EXPECT().
					CreateVehicle(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
			},
		},
		{
			name: "DuplicateLicensePlate",
			body: gin.H{
				"license_plate": vehicle.LicensePlate,
				"model":         vehicle.Model.String,
				"image_url":     vehicle.ImageUrl.String,
				"capacity":      vehicle.Capacity.Int32,
	},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
	},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.CreateVehicleParams{
					ID: vehicle.ID,
					DriverID: vehicle.DriverID,
					LicensePlate: vehicle.LicensePlate,
					Model: sql.NullString{String: vehicle.Model.String, Valid: true},
					ImageUrl: sql.NullString{String: vehicle.ImageUrl.String, Valid: true},
					Capacity: sql.NullInt32{Int32: vehicle.Capacity.Int32, Valid: true},
				}
				pqErr := &pq.Error{Code: "23505"} // unique_violation
				user.Role = string(util.RoleDriver)
				store.EXPECT().
					GetUserByID(gomock.Any(), gomock.Eq(user.ID)).
					Times(1).
					Return(user, nil)

				store.EXPECT().
					CreateVehicle(gomock.Any(), EqCreateVehicleParams(arg)).
					Times(1).
					Return(db.Vehicle{}, pqErr)
				},
					checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
					require.Equal(t, http.StatusConflict, recorder.Code)
				},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()

			data, err := json.Marshal(tc.body)
			require.NoError(t, err)

			url := "/vehicles/create"
			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
			require.NoError(t, err)
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func TestGetVehicle(t *testing.T) {
	user,_ := randomUser(t)
	vehicle := RandomVehicle(t)
	vehicle.DriverID = user.ID

	testCases := []struct {
		name string
		vehicleID string
		setupAuth func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			vehicleID: vehicle.ID.String(),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetVehicleByID(gomock.Any(), gomock.Eq(vehicle.ID)).
					Times(1).
					Return(vehicle, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
				requireBodyMatchVehicle(t, recorder.Body, vehicle)
			},
		},
		{
			name: "InternalServerError",
			vehicleID: vehicle.ID.String(),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetVehicleByID(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.Vehicle{}, sql.ErrConnDone)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusInternalServerError, recorder.Code)
				
			},
		},
		{
			name: "UnAutorizedUser",
			vehicleID: vehicle.ID.String(),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, uuid.New(), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetVehicleByID(gomock.Any(), gomock.Eq(vehicle.ID)).
					Times(1).
					Return(vehicle, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusUnauthorized, recorder.Code)
				
			},
		},
		{
			name: "NotFound",
			vehicleID: vehicle.ID.String(),
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, uuid.New(), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetVehicleByID(gomock.Any(), gomock.Eq(vehicle.ID)).
					Times(1).
					Return(db.Vehicle{}, sql.ErrNoRows)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusNotFound, recorder.Code)
				
			},
		},
		{
			name: "InvalidID",
			vehicleID: "invalid-uuid",
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, uuid.New(), time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				store.EXPECT().
					GetVehicleByID(gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusBadRequest, recorder.Code)
				
			},
		},
	}

	for i := range testCases {
		tc := testCases[i]

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/vehicles/%s", tc.vehicleID)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)		

		})
	}
}

func TestGetVehicleByDriversID(t *testing.T) {
	user, _ := randomUser(t)
	vehicles := make([]db.Vehicle, 10)
	for i:=0;i <10 ;i++ {
		vehicles[i] = RandomVehicle(t)
		vehicles[i].DriverID = user.ID
	}
	testCases := []struct{
		name string
		params GetVehiclesByDriverIDParams
		setupAuth func(t *testing.T, request *http.Request, tokenMaker token.Maker)
		buildStubs func(store *mockdb.MockStore)
		checkResponse func(t *testing.T, recorder *httptest.ResponseRecorder)
	}{
		{
			name: "OK",
			params: GetVehiclesByDriverIDParams{
			PageID: 1,
			PageSize: 5,
			},
			setupAuth: func(t *testing.T, request *http.Request, tokenMaker token.Maker) {
				addAuthorization(t, request, tokenMaker, authorizationTypeBearer, user.ID, time.Minute)
			},
			buildStubs: func(store *mockdb.MockStore) {
				arg := db.GetVehiclesByDriverIDParams{
					DriverID: user.ID,
					Limit: 5,
					Offset: 0,
				}
				store.EXPECT().
					GetVehiclesByDriverID(gomock.Any(), gomock.Eq(arg)).
					Times(1).
					Return(vehicles, nil)
			},
			checkResponse: func(t *testing.T, recorder *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusOK, recorder.Code)
			},
		},
	}
	for i := range testCases{
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			store := mockdb.NewMockStore(ctrl)
			tc.buildStubs(store)

			server := NewTestServer(t, store)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/vehicles?page_id=%d&page_size=%d", tc.params.PageID, tc.params.PageSize)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)	
			tc.setupAuth(t, request, server.tokenMaker)
			server.router.ServeHTTP(recorder, request)
			tc.checkResponse(t, recorder)
		})
	}
}

func requireBodyMatchVehicle(t *testing.T, body *bytes.Buffer, vehicle db.Vehicle) {
	data, err := io.ReadAll(body)
	require.NoError(t, err)

	var gotVehicle CreateVehicleResponse
	err = json.Unmarshal(data, &gotVehicle)
	require.NoError(t, err)

	require.Equal(t, vehicle.ID, gotVehicle.ID)
	require.Equal(t, vehicle.DriverID, gotVehicle.DriverID)
	require.Equal(t, vehicle.LicensePlate, gotVehicle.LicensePlate)
	require.Equal(t, vehicle.Model.String, gotVehicle.Model)
	require.Equal(t, vehicle.ImageUrl.String, gotVehicle.ImageUrl)
	require.Equal(t, vehicle.Capacity.Int32, gotVehicle.Capacity)
}