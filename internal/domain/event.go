package domain

const BookingConfirmedEventType = "booking.confirmed"

type BookingConfirmedPayload struct {
	BookingID int64 `json:"booking_id"`
	SeatID    int64 `json:"seat_id"`
	UserID    int64 `json:"user_id"`
}
