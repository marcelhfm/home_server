package types

import (
	"time"

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
	Id   int
	Name string
}

type DatasourceLastSeen struct {
	Datasource Datasource
	Last_seen  *time.Time
}
