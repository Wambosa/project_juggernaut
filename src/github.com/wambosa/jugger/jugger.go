package jugger
import "time"

type ReceivedMessage struct {
	ReceivedMessageId int
	ParcelTypeId int
	MessageText string
	UserId int
	JobId int
	ParseStatusId int
	CreatedOn time.Time
	LastUpdated time.Time
}

type ReceivedMessageMetadata struct {
	ReceivedMessageMetadataId int
	ReceivedMessageId int
	Key string
	Value string
}

type ParcelType struct {
	ParcelTypeId int
	ParcelTypeName string
}

type User struct {
	UserId int
	NickName string
	FirstName string
	LastName string
	LastUpdated time.Time
}