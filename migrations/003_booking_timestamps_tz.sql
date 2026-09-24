BEGIN;

ALTER TABLE bookings
ALTER COLUMN created_at TYPE TIMESTAMPTZ
USING created_at AT TIME ZONE 'Europe/Moscow';

ALTER TABLE bookings
ALTER COLUMN updated_at TYPE TIMESTAMPTZ
USING updated_at AT TIME ZONE 'Europe/Moscow';

ALTER TABLE bookings
ALTER COLUMN expires_at TYPE TIMESTAMPTZ
USING expires_at AT TIME ZONE 'Europe/Moscow';

COMMIT;