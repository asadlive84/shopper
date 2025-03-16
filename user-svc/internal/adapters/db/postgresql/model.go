package postgresql

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostgresDB struct {
	db *gorm.DB
}
type User struct {
	gorm.Model
	ID uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`

	Name              string         `gorm:"type:varchar(255);not null"` // Required field
	Email             string         `gorm:"uniqueIndex;not null"`       // Required field
	Password          string         `gorm:"not null"`                   // Required field
	PhoneNumber       *string        `gorm:"type:varchar(20)"`           // Nullable
	IsActive          bool           `gorm:"default:true"`               // Default value
	Age               *int32         `gorm:""`                           // Nullable
	Role              *int32         `gorm:""`                           // Nullable
	Permissions       *string        `gorm:"type:text"`                  // Nullable (JSON format or comma-separated)
	CreatedAt         time.Time      `gorm:"autoCreateTime"`             // Auto-generated
	UpdatedAt         time.Time      `gorm:"autoUpdateTime"`             // Auto-generated
	LastLogin         *time.Time     `gorm:""`                           // Nullable
	Status            *int32         `gorm:""`                           // Nullable
	ProfilePictureUrl *string        `gorm:"type:text"`                  // Nullable
	Metadata          *string        `gorm:"type:text"`                  // Nullable (Store as JSON string)
	IsDeleted         bool           `gorm:"default:false"`              // Default value
	DeletedAt         gorm.DeletedAt `gorm:"index"`                      // Soft delete
	TwoFactorEnabled  bool           `gorm:"default:false"`              // Default value
	TwoFactorSecret   *string        `gorm:""`                           // Nullable
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}
