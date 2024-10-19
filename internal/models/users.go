package models

type User struct {
	BaseModel
	Name     string `gorm:"type:varchar(100);not null;unique" json:"name"`
	Email    string `gorm:"type:varchar(256)" json:"email"`
	Password string `gorm:"type:varchar(256);not null" json:"password"`
	Type     int    `gorm:"type:int;default:1" json:"type"` // 1.内部（OA账号）
	Organize string `gorm:"type:varchar(256)" json:"organize"`
	Roles    []Role `gorm:"many2many:user_role;" json:"roles"`
}

type UserRole struct {
	UserID int `gorm:"primaryKey;index"` // UserID 是联合主键并定义索引
	RoleID int `gorm:"primaryKey"`       // RoleID 也是联合主键
}
