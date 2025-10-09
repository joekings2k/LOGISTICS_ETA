package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joekings2k/logistics-eta/util"
	"github.com/stretchr/testify/require"
)


func createRandomVehicle(t *testing.T, user User) Vehicle {

	user.Role = string(util.RoleDriver)
	arg := createVehicleParams{
		ID: 				uuid.New(),
		DriverID: 	user.ID,
		LicensePlate: util.RandomString(10),
		Model: sql.NullString{String: util.RandomString(6), Valid: true},
		ImageUrl: sql.NullString{String: util.RandomString(6), Valid: true},
		Capacity: sql.NullInt32{Int32: int32(util.RandomInt(1,100)), Valid: true},
	}

	vehicle, err := testQueries.createVehicle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, vehicle)

	require.Equal(t, arg.ID, vehicle.ID)
	require.Equal(t, arg.DriverID, vehicle.DriverID)
	require.Equal(t, arg.LicensePlate, vehicle.LicensePlate)
	require.Equal(t, arg.Model, vehicle.Model)
	require.Equal(t, arg.ImageUrl, vehicle.ImageUrl)
	require.Equal(t, arg.Capacity, vehicle.Capacity)

	require.NotZero(t, vehicle.CreatedAt)
	require.NotZero(t, vehicle.UpdatedAt)

	return vehicle
}


func TestCreateVehicle(t *testing.T) {
	user := createRandomUser(t)
	createRandomVehicle(t, user)
}

func TestGetVehicleByID(t *testing.T) {
	user := createRandomUser(t)
	vehicle1 := createRandomVehicle(t, user)

	vehicle2, err := testQueries.getVehicleByID(context.Background(), vehicle1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, vehicle2)

	require.Equal(t, vehicle1.ID, vehicle2.ID)
	require.Equal(t, vehicle1.DriverID, vehicle2.DriverID)
	require.Equal(t, vehicle1.LicensePlate, vehicle2.LicensePlate)
	require.Equal(t, vehicle1.Model, vehicle2.Model)
	require.Equal(t, vehicle1.ImageUrl, vehicle2.ImageUrl)
	require.Equal(t, vehicle1.Capacity, vehicle2.Capacity)

	require.WithinDuration(t, vehicle1.CreatedAt.Time, vehicle2.CreatedAt.Time,  time.Second)
	require.WithinDuration(t, vehicle1.UpdatedAt.Time, vehicle2.UpdatedAt.Time,  time.Second)
	
}

func TestGetVehicleByLicencePlate(t *testing.T) {
	user := createRandomUser(t)
	vehicle1 := createRandomVehicle(t, user)

	vehicle2, err := testQueries.getVehicleByLicensePlate(context.Background(), vehicle1.LicensePlate)
	require.NoError(t, err)
	require.NotEmpty(t, vehicle2)

	require.Equal(t, vehicle1.ID, vehicle2.ID)
	require.Equal(t, vehicle1.DriverID, vehicle2.DriverID)
	require.Equal(t, vehicle1.LicensePlate, vehicle2.LicensePlate)
	require.Equal(t, vehicle1.Model, vehicle2.Model)
	require.Equal(t, vehicle1.ImageUrl, vehicle2.ImageUrl)
	require.Equal(t, vehicle1.Capacity, vehicle2.Capacity)

	require.WithinDuration(t, vehicle1.CreatedAt.Time, vehicle2.CreatedAt.Time,  time.Second)
	require.WithinDuration(t, vehicle1.UpdatedAt.Time, vehicle2.UpdatedAt.Time,  time.Second)
	
}

func TestGetVehiclesByDriverID(t *testing.T) {
	user := createRandomUser(t)
	user.Role = string(util.RoleDriver)
	
	n := 5
	for i := 0; i < n; i++ {
		createRandomVehicle(t, user)
	}
	arg := getVehiclesByDriverIDParams{
		DriverID: user.ID,
		Limit: 3,
		Offset: 0,
	}
	vehicles, err := testQueries.getVehiclesByDriverID(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, vehicles)
	require.Len(t, vehicles, int(arg.Limit))
	for _, vehicle := range vehicles {
		require.NotEmpty(t, vehicle)
		require.Equal(t, user.ID, vehicle.DriverID)
		

	}

	arg = getVehiclesByDriverIDParams{
		DriverID: uuid.New(),
		Limit: 3,
		Offset: 0,
	}
	vehicles, err = testQueries.getVehiclesByDriverID(context.Background(), arg)
	require.NoError(t, err)
	require.Equal(t, len(vehicles), 0)

}

func TestUpdateVehicel(t *testing.T) {
	user := createRandomUser(t)
	vehicle1 := createRandomVehicle(t, user)

	arg := updateVehicleParams{
		ID: vehicle1.ID,
		Model: sql.NullString{String: util.RandomString(6), Valid: true},
		ImageUrl: sql.NullString{String: util.RandomString(6), Valid: true},
		Capacity: sql.NullInt32{Int32: int32(util.RandomInt(1,100)), Valid: true},
	}

	vehicle2, err := testQueries.updateVehicle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, vehicle2)
	
	require.Equal(t, vehicle1.ID, vehicle2.ID)
	require.Equal(t, vehicle1.DriverID, vehicle2.DriverID)
	require.Equal(t, vehicle1.LicensePlate, vehicle2.LicensePlate)
	require.Equal(t, arg.Model, vehicle2.Model)
	require.Equal(t, arg.ImageUrl, vehicle2.ImageUrl)
	require.Equal(t, arg.Capacity, vehicle2.Capacity)


	
}


func TestUpdateVehiclePartial(t *testing.T) {
	user := createRandomUser(t)
	vehicle1 := createRandomVehicle(t, user)
	
	arg := updateVehicleParams{
		ID: vehicle1.ID,
		Model: sql.NullString{String: util.RandomString(6), Valid: true},
	}

	vehicle2, err := testQueries.updateVehicle(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, vehicle2)
	
	require.Equal(t, vehicle1.ID, vehicle2.ID)
	require.Equal(t, vehicle1.DriverID, vehicle2.DriverID)
	require.Equal(t, vehicle1.LicensePlate, vehicle2.LicensePlate)
	require.Equal(t, arg.Model, vehicle2.Model)
	require.Equal(t, vehicle1.ImageUrl, vehicle2.ImageUrl)
	require.Equal(t, vehicle1.Capacity, vehicle2.Capacity)
}

func TestDeletevehicle(t *testing.T) {
	Vehicle1 := createRandomVehicle(t, createRandomUser(t))

	err := testQueries.deleteVehicle(context.Background(), Vehicle1.ID)
	require.NoError(t, err)

	vehicle2, err := testQueries.getVehicleByID(context.Background(), Vehicle1.ID)

	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, vehicle2)
}