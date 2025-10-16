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


func createRandomRoute(t  *testing.T, user *User, vehicle *Vehicle) Route {
	

	arg := CreateRouteParams{
		ID: uuid.New(),
		DriverID: user.ID,
		VehicleID: vehicle.ID,
		OriginAddress: sql.NullString{String: "123 Main st", Valid: true},
		OriginLat: 37.7749,
		OriginLng: -122.4194,
		DestinationAddress: sql.NullString{String: "456 Elm st", Valid: true},
		DestinationLat: 37.7849,
		DestinationLng: -122.4094,
		EstimatedDistanceKm: sql.NullFloat64{Float64: 5.0, Valid: true},
		EstimatedDurationMin: sql.NullFloat64{Float64: 15.0, Valid: true},
		Status: "pending",
	}

	route, err := testQueries.CreateRoute(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, route)

	require.Equal(t, arg.ID, route.ID)
	require.Equal(t, arg.DriverID, route.DriverID)
	require.Equal(t, arg.VehicleID, route.VehicleID)
	require.Equal(t, arg.OriginAddress, route.OriginAddress)
	require.Equal(t, arg.OriginLat, route.OriginLat)
	require.Equal(t, arg.OriginLng, route.OriginLng)
	require.Equal(t, arg.DestinationAddress, route.DestinationAddress)
	require.Equal(t, arg.DestinationLat, route.DestinationLat)
	require.Equal(t, arg.DestinationLng, route.DestinationLng)
	require.Equal(t, arg.EstimatedDistanceKm, route.EstimatedDistanceKm)
	require.Equal(t, arg.EstimatedDurationMin, route.EstimatedDurationMin)
	require.Equal(t, arg.Status, route.Status)

	require.NotZero(t, route.CreatedAt)
	require.NotZero(t, route.UpdatedAt)

	return route
}


func TestCreateRoute(t *testing.T) {
	user := createRandomUser(t)
	vehicle  := createRandomVehicle(t, user)
	createRandomRoute(t, &user, &vehicle)
}

func TestGetRouteByID(t *testing.T) {
	user := createRandomUser(t)
	vehicle  := createRandomVehicle(t, user)
	route1 := createRandomRoute(t, &user, &vehicle)

	route2, err := testQueries.GetRouteByID(context.Background(), route1.ID)
	require.NoError(t, err)
	require.NotEmpty(t, route2)

	require.Equal(t, route1.ID, route2.ID)
	require.Equal(t, route1.DriverID, route2.DriverID)
	require.Equal(t, route1.VehicleID, route2.VehicleID)
	require.Equal(t, route1.OriginAddress, route2.OriginAddress)
	require.Equal(t, route1.OriginLat, route2.OriginLat)
	require.Equal(t, route1.OriginLng, route2.OriginLng)
	require.Equal(t, route1.DestinationAddress, route2.DestinationAddress)
	require.Equal(t, route1.DestinationLat, route2.DestinationLat)
	require.Equal(t, route1.DestinationLng, route2.DestinationLng)
	require.Equal(t, route1.EstimatedDistanceKm, route2.EstimatedDistanceKm)
	require.Equal(t, route1.EstimatedDurationMin, route2.EstimatedDurationMin)
	require.Equal(t, route1.Status, route2.Status)

	require.WithinDuration(t, route1.CreatedAt.Time, route2.CreatedAt.Time, 0)
	require.WithinDuration(t, route1.UpdatedAt.Time, route2.UpdatedAt.Time, 0)
}

func TestDeleteRoute(t *testing.T) {
	user := createRandomUser(t)
	vehicle  := createRandomVehicle(t, user)
	route := createRandomRoute(t, &user, &vehicle)

	err := testQueries.DeleteRoute(context.Background(), route.ID)
	require.NoError(t, err)

	route2, err := testQueries.GetRouteByID(context.Background(), route.ID)
	require.Error(t, err)
	require.EqualError(t, err, sql.ErrNoRows.Error())
	require.Empty(t, route2)

}

func TestRoutesByDriverID(t *testing.T) {
	user := createRandomUser(t)
	vehicle := createRandomVehicle(t, user)
	for i := 0; i <5; i++ {
		createRandomRoute(t, &user, &vehicle)
	}

	arg := GetRoutesByDriverIDParams{
		DriverID: vehicle.DriverID,
		Limit: 3,
		Offset: 0,
	}
	routes, err := testQueries.GetRoutesByDriverID(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, routes)
	require.Len(t, routes, 3)

	for _, route := range routes {
		require.NotEmpty(t, route)
		require.Equal(t, vehicle.DriverID, route.DriverID)
	}
}

func TestListRoutesByDriverAndS(t *testing.T) {
	user:= createRandomUser(t)
	vehicle := createRandomVehicle(t, user)
	for i := 0 ;i < 10; i++ {
		createRandomRoute(t, &user, &vehicle)
	}

	arg := ListRoutesByDriverAndStatusParams{
		DriverID: vehicle.DriverID,
		Status: string(util.RoutePending),
		Limit: 5,
		Offset: 0,
	}
	routes, err := testQueries.ListRoutesByDriverAndStatus(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, routes)
	require.Len(t, routes, 5)

	for _, route := range routes {
		require.NotEmpty(t, route)
		require.Equal(t, vehicle.DriverID, route.DriverID)
		require.Equal(t, string(util.RoutePending), route.Status)
	}
	
}

func TestUpdateRouteStatus(t *testing.T) {
	user := createRandomUser(t)
	vehicle := createRandomVehicle(t, user)
	route := createRandomRoute(t , &user, &vehicle)

	newStatus := util.RouteInProgress

	arg := UpdateRouteStatusParams{
		ID: route.ID,
		Status: string(newStatus),
	}

	route2, err := testQueries.UpdateRouteStatus(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, route2)

	require.Equal(t, route.ID, route2.ID)
	require.Equal(t, route.DriverID, route2.DriverID)
	require.Equal(t, route.VehicleID, route2.VehicleID)
	require.Equal(t, string(newStatus), route2.Status)
}


func TestUpdateRouteActualDuration(t *testing.T) {
	user := createRandomUser(t)
	vehicle := createRandomVehicle(t, user)
	route := createRandomRoute(t , &user, &vehicle)

	durationMinutes := float64((20 * time.Minute).Minutes())

	newActualDuration := sql.NullFloat64{Float64:durationMinutes, Valid: true}

	arg := UpdateRouteActualDurationParams{
		ID: route.ID,
		ActualDurationMin: newActualDuration,
	}

	route2, err := testQueries.UpdateRouteActualDuration(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, route2)

	require.Equal(t, route.ID, route2.ID)
	require.Equal(t, route.DriverID, route2.DriverID)
	require.Equal(t, route.VehicleID, route2.VehicleID)
	require.Equal(t, newActualDuration, route2.ActualDurationMin)
}
