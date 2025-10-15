func ({{.upperStartCamelObject}}) TableName() string {
    return {{.table}}
}

func new{{.upperStartCamelObject}}Model(db *gorm.DB{{if .withCache}}, redisConf gormc.RedisConfig, cacheExpiry time.Duration{{end}}) (*default{{.upperStartCamelObject}}Model, error) {
	{{if .withCache}}cachedConn, err := gormc.NewConn(db, redisConf, cacheExpiry)
	if err != nil {
		return nil, err
	}
	return &default{{.upperStartCamelObject}}Model{
		CachedConn: cachedConn,
		table: {{.table}},
	}, nil{{else}}return &default{{.upperStartCamelObject}}Model{
		conn: db,
		table: {{.table}},
	}, nil{{end}}
}
