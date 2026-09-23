package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id          uint   `json:"id" gorm:"primaryKey;not null;unique"`
	Username    string `json:"username" gorm:"not null"`
	Bank        string `json:"bank"`
	PhoneNumber string `json:"phone_number"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Tokens []Token       `json:"-" gorm:"foreignKey:user_id;references:id;constraint:OnDelete:CASCADE"`
	Events []EventMember `json:"-" gorm:"foreignKey:user_id;references:id;constraint:OnDelete:CASCADE"`
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
	Id        uint       `json:"id" gorm:"primaryKey;autoIncrement;not null;unique"`
	Name      string     `json:"name"`
	Ended     bool       `json:"ended" gorm:"default:false"`
	StartsAt  *time.Time `json:"starts_at" gorm:"index:events_reminder_idx,where:chat_id IS NOT NULL"` // nil — без даты
	ChatId    *int64     `json:"chat_id" gorm:"index"`                                                 // чат, где событие создал бот; nil — из мини-аппа
	CreatedAt time.Time
	UpdatedAt time.Time

	Members   []EventMember   `gorm:"foreignKey:event_id;references:id;constraint:OnDelete:CASCADE"`
	Products  []Product       `gorm:"foreignKey:event_id;references:id;constraint:OnDelete:CASCADE"`
	Invites   []Invite        `gorm:"foreignKey:event_id;references:id;constraint:OnDelete:CASCADE"`
	Reminders []EventReminder `json:"-" gorm:"foreignKey:event_id;references:id;constraint:OnDelete:CASCADE"`
}

// Отправленное ботом напоминание: по строке на событие и окно (24 или 3 часа).
type EventReminder struct {
	EventId     uint      `gorm:"primaryKey;autoIncrement:false"`
	WindowHours int       `gorm:"primaryKey;autoIncrement:false"`
	SentAt      time.Time `gorm:"not null;default:now()"`
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
	Price      float64 `json:"price" gorm:"default:0"`
	Bought     bool    `json:"bought" gorm:"default:false"`
	Equal      bool    `json:"equal" gorm:"default:true"`
	CreatedAt  time.Time
	UpdatedAt  time.Time

	Event    Event           `json:"-" gorm:"foreignKey:EventId;references:Id"`
	Consumer User            `json:"-" gorm:"foreignKey:ConsumerId;references:Id"`
	Members  []ProductMember `gorm:"foreignKey:ProductId;references:Id;constraint:OnDelete:CASCADE"`
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

	User    User    `gorm:"foreignKey:UserId;references:Id"`
	Product Product `json:"-" gorm:"foreignKey:ProductId;references:Id"`
}

type UpdateUser struct {
	// Id          uint    `json:"id" gorm:"primaryKey;not null;unique"`
	Bank        *string `json:"bank"`
	PhoneNumber *string `json:"phone_number"`
}

type CreateEvent struct {
	Name     string     `json:"name"`
	StartsAt *time.Time `json:"starts_at"`
}

type UpdateEvent struct {
	Id       uint       `json:"id" gorm:"primaryKey;not null;unique"`
	Name     *string    `json:"name"`
	Ended    *bool      `json:"ended" gorm:"default:false"`
	StartsAt *time.Time `json:"starts_at"`
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

type ProductList struct {
	Id     uint    `json:"id"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
	Count  int     `json:"count"`
	Bought bool    `json:"bought"`
}

type CreateProductMembers struct {
	UserId uint     `json:"user_id"`
	Price  *float64 `json:"price,omitempty"`
}

type CreateProduct struct {
	Name       string  `json:"name"`
	ConsumerId *uint   `json:"consumer_id"`
	Price      float64 `json:"price"`
	Bought     bool    `json:"bought"`
	Equal      bool    `json:"equal"`

	Members []CreateProductMembers `json:"members"`
}

type ProductResponse struct {
	Id               uint                    `json:"id"`
	Name             string                  `json:"name"`
	Price            float64                 `json:"price"`
	Bought           bool                    `json:"bought"`
	Equal            bool                    `json:"equal"`
	ConsumerId       uint                    `json:"consumer_id"`
	ConsumerUsername string                  `json:"consumer_username"`
	Members          []ProductMemberResponse `json:"members"`
}

type ProductMemberResponse struct {
	UserId   uint    `json:"user_id"`
	Username string  `json:"username"`
	Price    float64 `json:"price"`
}

type UpdateProduct struct {
	Name       string                `json:"name"`
	ConsumerId *uint                 `json:"consumer_id"`
	Price      float64               `json:"price"`
	Bought     bool                  `json:"bought"`
	Equal      bool                  `json:"equal"`
	Members    []UpdateProductMember `json:"members"`
}

type UpdateProductMember struct {
	UserId uint     `json:"user_id"`
	Price  *float64 `json:"price,omitempty"`
}

type Bill struct {
	ProductId uint    `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	CId       uint    `json:"consumer_id"`
	CName     string  `json:"consumer_name"`
	CPhone    string  `json:"consumer_phone"`
	CBank     string  `json:"consumer_bank"`
}
