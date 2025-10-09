CREATE TABLE vehicles (
    id UUID PRIMARY KEY,
    driver_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    license_plate TEXT NOT NULL,
    model TEXT,
    image_url TEXT,
    capacity INTEGER,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_vehicles_driver_id ON vehicles(driver_id);
CREATE UNIQUE INDEX idx_vehicles_license_plate ON vehicles(license_plate);

