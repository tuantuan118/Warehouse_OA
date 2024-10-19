package config

type MysqlConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	DbName   string `json:"db_name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type JWTConfig struct {
	SigningKey string `json:"signing_key"`
}

type ServerConfig struct {
	JWTInfo   JWTConfig   `json:"jwt"`
	MysqlInfo MysqlConfig `json:"mysql"`
}
