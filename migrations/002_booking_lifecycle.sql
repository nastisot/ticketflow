BEGIN;

ALTER TABLE bookings
ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'confirmed'
    CONSTRAINT bookings_status_check
    CHECK (status IN ('pending', 'confirmed', 'cancelled', 'expired'));

ALTER TABLE bookings
ALTER COLUMN status SET DEFAULT 'pending';

ALTER TABLE bookings
ADD COLUMN expires_at TIMESTAMP NULL;

ALTER TABLE bookings
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE bookings
DROP CONSTRAINT bookings_seat_id_key;

CREATE UNIQUE INDEX bookings_active_seat_unique
ON bookings (seat_id)
WHERE status IN ('pending', 'confirmed');

COMMIT;