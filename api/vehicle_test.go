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