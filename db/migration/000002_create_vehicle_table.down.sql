DROP INDEX IF EXISTS idx_vehicles_driver_id;
DROP INDEX IF EXISTS idx_vehicles_license_plate;

-- Drop the table
DROP TABLE IF EXISTS vehicles CASCADE;