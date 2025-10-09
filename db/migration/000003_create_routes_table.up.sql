CREATE TABLE routes (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    vehicle_id UUID NOT NULL REFERENCES vehicles(id) ON DELETE SET NULL,

    -- Geo fields
    origin_lat DOUBLE PRECISION NOT NULL,
    origin_lng DOUBLE PRECISION NOT NULL,
    destination_lat DOUBLE PRECISION NOT NULL,
    destination_lng DOUBLE PRECISION NOT NULL,

    -- Optional human-readable addresses
    origin_address TEXT,
    destination_address TEXT,

    -- Distance and time estimation fields
    estimated_distance_km DOUBLE PRECISION,
    estimated_duration_min DOUBLE PRECISION,
    actual_duration_min DOUBLE PRECISION,

    -- Status: e.g. "pending", "in_progress", "completed", "cancelled"
    status TEXT NOT NULL,

    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes for performance
CREATE INDEX idx_routes_driver_id ON routes(driver_id);
CREATE INDEX idx_routes_status ON routes(status);
CREATE INDEX idx_routes_vehicle_id ON routes(vehicle_id);
CREATE INDEX idx_routes_origin_dest ON routes(origin_lat, origin_lng, destination_lat, destination_lng);