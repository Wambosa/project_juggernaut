package jugger
import "time"

type ReceivedMessage struct {
	ReceivedMessage int
	ParcelTypeId int
	MessageText string
	UserId int
	JobId int
	ParseStatusId int
	CreatedOn time.Time
	LastUpdated time.Time
}