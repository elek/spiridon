package db

import (
	"context"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm/logger"
	"time"
)

type DBLog struct {
}

func (d *DBLog) LogMode(level logger.LogLevel) logger.Interface {
	return d
}

func (d *DBLog) Info(ctx context.Context, s string, i ...interface{}) {
	log.Info().Msgf(s, i...)
}

func (d *DBLog) Warn(ctx context.Context, s string, i ...interface{}) {
	log.Warn().Msgf(s, i...)
}

func (d *DBLog) Error(ctx context.Context, s string, i ...interface{}) {
	log.Error().Msgf(s, i...)
}

func (d *DBLog) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	sql, _ := fc()
	log.Debug().Msg(sql)
}
