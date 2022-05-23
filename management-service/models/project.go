package models

import (
	"time"

	"github.com/gojek/xp/common/api/schema"
)

type Project struct {
	CreatedAt        time.Time `json:"created_at"`
	Id               int64     `json:"id"`
	RandomizationKey string    `json:"randomization_key"`
	Segmenters       []string  `json:"segmenters"`
	UpdatedAt        time.Time `json:"updated_at"`
	Username         string    `json:"username"`
}

type ProjectSettings struct {
	CreatedAt            time.Time         `json:"created_at"`
	EnableS2idClustering bool              `json:"enable_s2id_clustering"`
	Passkey              string            `json:"passkey"`
	ProjectId            int64             `json:"project_id"`
	RandomizationKey     string            `json:"randomization_key"`
	Segmenters           ProjectSegmenters `json:"segmenters"`
	UpdatedAt            time.Time         `json:"updated_at"`
	Username             string            `json:"username"`
}

func (c *Project) ToApiSchema() schema.Project {
	project := schema.Project{
		Id:               c.Id,
		CreatedAt:        c.CreatedAt,
		RandomizationKey: c.RandomizationKey,
		Segmenters:       c.Segmenters,
		UpdatedAt:        c.UpdatedAt,
		Username:         c.Username,
	}

	return project
}
