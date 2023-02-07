package model

type EventSearchResultEntered struct {
	ID             string
	BonjourID      string
	Query          string
	ResultPosition uint32
	Destination    string
}
