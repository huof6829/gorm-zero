func ({{.upperStartCamelObject}}) TableName() string {
    return strings.Trim({{.table}}, "`")
}

func new{{.upperStartCamelObject}}Model(db *gorm.DB{{if .withCache}}, cache *gormc.RedisCache{{end}}) *default{{.upperStartCamelObject}}Model {
	{{if .withCache}}cachedConn := gormc.NewConnWithCache(db, cache)
	return &default{{.upperStartCamelObject}}Model{
		CachedConn: cachedConn,
		table: strings.Trim({{.table}}, "`"),
	}{{else}}
	return &default{{.upperStartCamelObject}}Model{
		conn: db,
		table: strings.Trim({{.table}}, "`"),
	}{{end}}
}
