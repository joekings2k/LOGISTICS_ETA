package util

type Role string
type RouteStatus string

const (
	RoleAdmin 		Role = "admin"
	RoleDriver 		Role = "driver"
	RoleCustomer 	Role = "customer"
)

const (
	RoutePending 		RouteStatus = "pending"
	RouteInProgress RouteStatus = "in_progress"
	RouteCompleted 	RouteStatus = "completed"
	RouteCancelled 	RouteStatus = "cancelled"
)