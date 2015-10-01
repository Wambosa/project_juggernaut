package jugger
import "time"

type Mind struct {
	MindId int
	MindName string
	Nosiness int
	Sassyness int
	UniqueAddress string
	LastUpdated time.Time
}

type MindCapability struct {
	MindCapabilityId int
	MindId int
	ActionId int
}

type ReceivedMessage struct {
	ReceivedMessageId int
	ParcelTypeId int
	MessageText string
	UserId int
	JobId int
	ParseStatusId int
	CreatedOn time.Time
	LastUpdated time.Time
	Metadata []ReceivedMessageMetadata
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