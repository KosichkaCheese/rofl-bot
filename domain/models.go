package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id          uint   `json:"id" gorm:"primaryKey;not null;unique"`
	Username    string `json:"username" gorm:"not null;unique"`
	Bank        string `json:"bank"`
	PhoneNumber string `json:"phone_number"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Tokens []Token       `gorm:"foreignKey:user_id;references:id;constraint:OnDelete:CASCADE"`
	Events []EventMember `gorm:"foreignKey:user_id;references:id;constraint:OnDelete:CASCADE"`
}

type Token struct {
	Refresh   string    `json:"refresh" gorm:"primaryKey;not null;unique"`
	UserId    uint      `json:"user_id" gorm:"not null;index"`
	Revoked   bool      `json:"revoke" gorm:"default:false"`
	ExpiresIn time.Time `json:"expires_in"`
	CreatedAt time.Time
	UpdatedAt time.Time

	User User `gorm:"foreignKey:user_id;references:id"`
}

type Event struct {
	Id        uint   `json:"id" gorm:"primaryKey;autoIncrement;not null;unique"`
	Name      string `json:"name"`
	Ended     bool   `json:"ended" gorm:"default:false"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Members  []EventMember `gorm:"foreignKey:event_id;references:id;constraint:OnDelete:CASCADE"`
	Products []Product     `gorm:"foreignKey:event_id;references:id;constraint:OnDelete:CASCADE"`
	Invites  []Invite      `gorm:"foreignKey:event_id;references:id;constraint:OnDelete:CASCADE"`
}

type Role struct {
	Id   uint   `json:"id" gorm:"primaryKey;autoIncrement;not null;unique"`
	Name string `json:"name"`

	Members []EventMember `gorm:"foreignKey:role_id;references:id"`
}

type EventMember struct {
	Id      uint `json:"id" gorm:"primaryKey;autoIncrement;not null;unique"`
	UserId  uint `json:"user_id" gorm:"not null;index"`
	EventId uint `json:"event_id" gorm:"not null;index"`
	RoleId  uint `json:"role_id" gorm:"not null"`

	User  User  `gorm:"foreignKey:user_id;references:id"`
	Event Event `gorm:"foreignKey:event_id;references:id"`
	Role  Role  `gorm:"foreignKey:role_id;references:id"`
}

type Product struct {
	Id         uint    `json:"id" gorm:"primaryKey;autoIncrement;not null;unique"`
	Name       string  `json:"name"`
	EventId    uint    `json:"event_id" gorm:"not null;index"`
	ConsumerId uint    `json:"consumer_id" gorm:"not null"`
	Amount     uint    `json:"amount" gorm:"default:0"`
	Price      float64 `json:"price" gorm:"default:0"`
	Bought     bool    `json:"bought" gorm:"default:false"`
	Equal      bool    `json:"equal" gorm:"default:true"`
	CreatedAt  time.Time
	UpdatedAt  time.Time

	Event    Event           `gorm:"foreignKey:event_id;references:id"`
	Consumer User            `gorm:"foreignKey:consumer_id;references:id"`
	Members  []ProductMember `gorm:"foreignKey:product_id;references:id;constraint:OnDelete:CASCADE"`
}

type Invite struct {
	Id        uuid.UUID `json:"id" gorm:"primaryKey;not null;unique"`
	EventId   uint      `json:"event_id" gorm:"not null;index"`
	ExpiresAt time.Time `json:"expires_at"`

	Event Event `gorm:"foreignKey:event_id;references:id"`
}

type ProductMember struct {
	Id        uint    `json:"id" gorm:"primaryKey;autoIncrement;not null;unique"`
	UserId    uint    `json:"user_id" gorm:"not null;index"`
	ProductId uint    `json:"product_id" gorm:"not null;index"`
	Price     float64 `json:"amount" gorm:"default:0"`

	User    User    `gorm:"foreignKey:user_id;references:id"`
	Product Product `gorm:"foreignKey:product_id;references:id"`
}

type UpdateUser struct {
	// Id          uint    `json:"id" gorm:"primaryKey;not null;unique"`
	Bank        *string `json:"bank"`
	PhoneNumber *string `json:"phone_number"`
}

type CreateEvent struct {
	Name string `json:"name"`
}

type UpdateEvent struct {
	Id    uint    `json:"id" gorm:"primaryKey;not null;unique"`
	Name  *string `json:"name"`
	Ended *bool   `json:"ended" gorm:"default:false"`
}

type EventMembers struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	RoleID   uint   `json:"role_id"`
	RoleName string `json:"role_name"`
}

type EventWithMembersCount struct {
	Event `json:"event" gorm:"embedded"`
	Count int `json:"count"`
}

type EventWithAdmin struct {
	Event   EventWithMembersCount `json:"event" gorm:"embedded"`
	IsAdmin bool                  `json:"is_admin"`
}
