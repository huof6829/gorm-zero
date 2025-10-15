package {{.pkg}}
{{if .withCache}}
import (
	"time"
	"github.com/SpectatorNan/gorm-zero/gormc"
	"gorm.io/gorm"
)
{{else}}
import (
	"gorm.io/gorm"
	{{ if or (.gormCreatedAt) (.gormUpdatedAt) }} "time" {{ end }}
)
{{end}}
var _ {{.upperStartCamelObject}}Model = (*custom{{.upperStartCamelObject}}Model)(nil)

type (
	// {{.upperStartCamelObject}}Model is an interface to be customized, add more methods here,
	// and implement the added methods in custom{{.upperStartCamelObject}}Model.
	{{.upperStartCamelObject}}Model interface {
		{{.lowerStartCamelObject}}Model
		custom{{.upperStartCamelObject}}LogicModel
	}

	custom{{.upperStartCamelObject}}Model struct {
		*default{{.upperStartCamelObject}}Model
	}

	custom{{.upperStartCamelObject}}LogicModel interface {

    	}
)
{{ if or (.gormCreatedAt) (.gormUpdatedAt) }}
// BeforeCreate hook create time
func (s *{{.upperStartCamelObject}}) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	{{ if .gormCreatedAt }}s.CreatedAt = now{{ end }}
	{{ if .gormUpdatedAt}}s.UpdatedAt = now{{ end }}
	return nil
}
{{ end }}
{{ if .gormUpdatedAt}}
// BeforeUpdate hook update time
func (s *{{.upperStartCamelObject}}) BeforeUpdate(tx *gorm.DB) error {
	s.UpdatedAt = time.Now()
	return nil
}
{{ end }}
// New{{.upperStartCamelObject}}Model returns a model for the database table.
// For cached models, provide redisConf and cacheExpiry parameters.
// 
// Single node example:
//   redisConf := gormc.RedisConfig{Addr: "127.0.0.1:6379", DB: 0}
//   model, err := New{{.upperStartCamelObject}}Model(db, redisConf, time.Hour)
//
// Cluster example:
//   redisConf := gormc.RedisConfig{ClusterAddrs: []string{"node1:6379", "node2:6379", "node3:6379"}}
//   model, err := New{{.upperStartCamelObject}}Model(db, redisConf, time.Hour)
func New{{.upperStartCamelObject}}Model(conn *gorm.DB{{if .withCache}}, redisConf gormc.RedisConfig, cacheExpiry time.Duration{{end}}) ({{.upperStartCamelObject}}Model, error) {
	{{if .withCache}}defaultModel, err := new{{.upperStartCamelObject}}Model(conn, redisConf, cacheExpiry)
	if err != nil {
		return nil, err
	}
	return &custom{{.upperStartCamelObject}}Model{
		default{{.upperStartCamelObject}}Model: defaultModel,
	}, nil{{else}}defaultModel, err := new{{.upperStartCamelObject}}Model(conn)
	if err != nil {
		return nil, err
	}
	return &custom{{.upperStartCamelObject}}Model{
		default{{.upperStartCamelObject}}Model: defaultModel,
	}, nil{{end}}
}
{{if .withCache}}

func (m *default{{.upperStartCamelObject}}Model) customCacheKeys(data *{{.upperStartCamelObject}}) []string {
    if data == nil {
        return []string{}
    }
	return []string{}
}
{{ end }}
