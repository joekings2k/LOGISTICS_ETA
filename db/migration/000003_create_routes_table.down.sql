DROP INDEX IF EXISTS idx_routes_driver_id;
DROP INDEX IF EXISTS idx_routes_status;
DROP INDEX IF EXISTS idx_routes_vehicle_id;
DROP INDEX IF EXISTS idx_routes_origin_dest;

DROP TABLE IF EXISTS routes CASCADE;