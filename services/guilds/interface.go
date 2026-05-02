package guilds

import (
	"context"

	"gorm.io/gorm"
)

var (
	Service GuildsService
)

type GuildsService interface {
	ResetGuild(ctx context.Context, guildID string) error
}

type GuildsServiceImpl struct {
	db *gorm.DB
}

func Init(db *gorm.DB) {
	if Service == nil {
		Service = &GuildsServiceImpl{db: db}
	}
}
