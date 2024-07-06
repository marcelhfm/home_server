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
	Message string
	Color   string
}
