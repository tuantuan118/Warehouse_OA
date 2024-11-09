package models

type Order struct {
	BaseModel
	ProduceId     int     `gorm:"-" json:"produceId"`
	OrderNumber   string  `gorm:"type:varchar(256);not null" json:"orderNumber"`
	Name          string  `gorm:"type:varchar(256);not null" json:"name"`
	Specification string  `gorm:"type:varchar(256)" json:"specification"`
	Price         float64 `gorm:"type:decimal(10,2)" json:"price"`
	Amount        int     `gorm:"type:int(11);not null" json:"amount"`
	TotalPrice    float64 `gorm:"type:decimal(10,2)" json:"totalPrice"`
	FinishPrice   float64 `gorm:"type:decimal(10,2)" json:"finishPrice"`
	UnFinishPrice float64 `gorm:"type:decimal(10,2)" json:"unFinishPrice"`
	Status        int     `gorm:"type:int(11);not null" json:"status"`
	CustomerName  string  `gorm:"type:varchar(256)" json:"customerName"`
	Salesman      string  `gorm:"type:varchar(256)" json:"salesman"`
	UserList      []User  `gorm:"many2many:order_user;" json:"userList"`
}
