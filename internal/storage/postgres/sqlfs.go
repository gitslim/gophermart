package postgres

import (
	"embed"
	"log"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed sql/*.sql
var sqlFS embed.FS

// loadQueries загружает SQL-запросы из файлов и присваивает их переменным.
func loadQueries(queries map[string]*string) {
	for file, qPtr := range queries {
		data, err := sqlFS.ReadFile(filepath.Join("sql", file))
		if err != nil {
			log.Fatalf("Ошибка загрузки SQL-запроса из файла %s: %v", file, err)
		}
		*qPtr = string(data)
	}
}
