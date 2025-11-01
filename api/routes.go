package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	db "github.com/joekings2k/logistics-eta/db/sqlc"
	"github.com/joekings2k/logistics-eta/token"
	"github.com/joekings2k/logistics-eta/util"
)



type CreateRoutesRequest struct{
	VehicleID string `json:"vehicle_id" binding:"required"`
	OriginAddress string `json:"origin_address" binding:"required"`
	OriginLat float64 `json:"origin_lat" binding:"required"`
	OriginLng float64 `json:"origin_lng" binding:"required"`
	DestinationAddress string `json:"destination_address" binding:"required"`
	DestinationLat float64 `json:"destination_lat" binding:"required"`
	DestinationLng float64 `json:"destination_lng" binding:"required"`

}

type CreateRoutesResponse struct{
	ID uuid.UUID `json:"id"`
	VehicleID string `json:"vehicle_id"`
	OriginAddress string `json:"origin_address"`
	OriginLat float64 `json:"origin_lat"`
	OriginLng float64 `json:"origin_lng"`
	DestinationAddress string `json:"destination_address"`
	DestinationLat float64 `json:"destination_lat"`
	DestinationLng float64 `json:"destination_lng"`
}

func newRoutesResponse(route db.Route) CreateRoutesResponse {
	return CreateRoutesResponse{
		ID: route.ID,
		VehicleID: route.VehicleID.String(),
		OriginAddress: route.OriginAddress.String,
		OriginLat: route.OriginLat,
		OriginLng: route.OriginLng,
		DestinationAddress: route.DestinationAddress.String,
		DestinationLat: route.DestinationLat,
		DestinationLng: route.DestinationLng,
	}
}

func (server *Server) CreateRoute(ctx *gin.Context){

	var req CreateRoutesRequest
	if err := ctx.ShouldBindJSON(&req);err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
	}

	authPayload := ctx.MustGet(authorizationPayloadKey).(*token.Payload)
	vehicleID,err := uuid.Parse(req.VehicleID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}

	vehicle, err := server.store.GetVehicleByID(ctx, vehicleID)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	if !util.EnsureOwnership(ctx, vehicle.DriverID, authPayload.UserID) {
		return
	}

	arg := db.CreateRouteParams{
		ID: uuid.New(),
		DriverID: vehicle.DriverID,
		VehicleID: vehicleID,
		OriginAddress: sql.NullString{String: req.OriginAddress, Valid: true},
		OriginLat: req.OriginLat,
		OriginLng: req.OriginLng,
		DestinationAddress: sql.NullString{String: req.DestinationAddress, Valid: true},
		DestinationLat: req.DestinationLat,
		DestinationLng: req.DestinationLng,
	}

	route, err := server.store.CreateRoute(ctx, arg)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, newRoutesResponse(route))	
}

type GetRouteRequest struct{
	ID string `uri:"id" binding:"required"`
}
func (server *Server) GetRoute(ctx *gin.Context) {
 var req GetRouteRequest
 if err := ctx.ShouldBindUri(&req);err != nil{
	 ctx.JSON(http.StatusBadRequest, errorResponse(err))
	 return
 }

 routeID, err := uuid.Parse(req.ID)
 if err != nil {
	 ctx.JSON(http.StatusBadRequest, errorResponse(err))
	 return
 }

 route, err := server.store.GetRouteByID(ctx, routeID)
 if err != nil {
	 if err == sql.ErrNoRows {
		 ctx.JSON(http.StatusNotFound, errorResponse(err))
		 return
	 }
	 ctx.JSON(http.StatusInternalServerError, errorResponse(err))
	 return
 }

 ctx.JSON(http.StatusOK, gin.H{"route": route})
}
type CompleteRouteRequest struct{
	ID string `uri:"id" binding:"required"`
}

func (server *Server) CompleteRoute(ctx *gin.Context) {
	var req CompleteRouteRequest
	if err := ctx.ShouldBindUri(&req); err != nil{
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	routeID , err := uuid.Parse(req.ID)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, errorResponse(err))
		return
	}
	arg := db.CompleteRouteParams{
		ID: routeID,
		Status: string(util.RouteCompleted),
		ActualDurationMin: sql.NullFloat64{Float64: float64((20 * time.Minute).Minutes()), Valid: true},
	}

	route, err := server.store.CompleteRoute(ctx, arg)
	if err != nil {
		if err == sql.ErrNoRows {
			ctx.JSON(http.StatusNotFound, errorResponse(err))
			return
		}
		ctx.JSON(http.StatusInternalServerError, errorResponse(err))
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"route": route})


}


