package types

import (
	"github.com/google/uuid"
)

type CommandBody struct {
	Command string
}

type CommandRequest struct {
	Id           uuid.UUID
	Command      int
	DatasourceId int
}

type CommandResponse struct {
	Id           uuid.UUID
	Command      int
	DatasourceId int
	Error        error
}

type Datasource struct {
	Id        int
	Name      string
	Status    string
	Last_seen string
}

type FormattedLogs struct {
	Message   string
	Timestamp string
	Level     string
	Color     string
}

type DsDataPaneProps struct {
	DsId          string
	DsName        string
	Chart         string
	Co2           int
	Moisture      int
	DisplayStatus int
	LastSeen      string
	Status        string
	Data          bool
	DsType        string
}
