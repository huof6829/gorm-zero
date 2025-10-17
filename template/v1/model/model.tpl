package {{.pkg}}
{{if .withCache}}
import (
	"github.com/huof6829/gorm-zero/gormc"
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
func New{{.upperStartCamelObject}}Model(conn *gorm.DB{{if .withCache}}, cache *gormc.RedisCache{{end}}) {{.upperStartCamelObject}}Model {
	{{if .withCache}}defaultModel := new{{.upperStartCamelObject}}Model(conn, cache)
	return &custom{{.upperStartCamelObject}}Model{
		default{{.upperStartCamelObject}}Model: defaultModel,
	}{{else}}defaultModel := new{{.upperStartCamelObject}}Model(conn)
	return &custom{{.upperStartCamelObject}}Model{
		default{{.upperStartCamelObject}}Model: defaultModel,
	}{{end}}
}
{{if .withCache}}

func (m *default{{.upperStartCamelObject}}Model) customCacheKeys(data *{{.upperStartCamelObject}}) []string {
    if data == nil {
        return []string{}
    }
	return []string{}
}
{{ end }}
