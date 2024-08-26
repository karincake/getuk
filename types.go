package getuk

import (
	"time"

	"gorm.io/gorm"
)

type Pagination struct {
	PageNumber  int  `json:"page-number"`
	PageSize    int  `json:"page-size"`
	PageNoLimit bool `json:"page-no-limit"`
}

type DateModel struct {
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `json:"deletedAt,omitempty" gorm:"index"`
}

type Mode string
type FlatJoinOpt struct {
	Src      string
	SrcFkCol string
	Ref      string
	RefCol   string
	Mode     Mode
	Clause   string
	Prefix   string
	Cols     []string
}
