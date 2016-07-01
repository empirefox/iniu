package conf

import (
	"time"

	"github.com/mcuadros/go-defaults"
)

const (
	DAY         = time.Duration(24) * time.Hour
	C_BUCKET    = "bucket"
	MONTH_COUNT = 12
	IMG_PRE_FMT = "200601-"
)

var (
	Defaults DefaultsContainer
)

type DefaultsContainer struct {
	PageSize int `default:"20"`
}

func init() {
	defaults.SetDefaults(&Defaults)
}
