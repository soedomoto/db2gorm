package v2

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	dataloadergen "github.com/soedomoto/db2gorm/module/dataloader"
	"github.com/soedomoto/db2gorm/properties"

	"gopkg.in/yaml.v2"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gen"
	"gorm.io/gorm"
)

type DBType string

const (
	dbMySQL     DBType = "mysql"
	dbPostgres  DBType = "postgres"
	dbSQLite    DBType = "sqlite"
	dbSQLServer DBType = "sqlserver"
)

func ConnectDB(t DBType, dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn cannot be empty")
	}

	switch t {
	case dbMySQL:
		return gorm.Open(mysql.Open(dsn))
	case dbPostgres:
		return gorm.Open(postgres.Open(dsn))
	case dbSQLite:
		return gorm.Open(sqlite.Open(dsn))
	case dbSQLServer:
		return gorm.Open(sqlserver.Open(dsn))
	default:
		return nil, fmt.Errorf("unknow db %q (support mysql || postgres || sqlite || sqlserver for now)", t)
	}
}

// =================================================================================

func LoadConfigFile(path string) (*properties.YamlProperties, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close() // nolint
	var props properties.YamlProperties
	if cmdErr := yaml.NewDecoder(file).Decode(&props); cmdErr != nil {
		return nil, cmdErr
	}

	return &props, nil
}

// =================================================================================

type migrator struct{ *gorm.DB }
type Table struct {
	gorm.TableType
	TableName string
}
type Column struct {
	gorm.ColumnType
	TableName string
}

func (t *migrator) GetTables() (result []string, err error) {
	tableList, err := t.Migrator().GetTables()
	if err != nil {
		return nil, err
	}

	return tableList, nil
}

func (t *migrator) GetTableType(tableName string) (tableType gorm.TableType, err error) {
	tableType, err = t.Migrator().TableType(tableName)
	return tableType, err
}

func (t *migrator) GetTableColumns(tableName string) (result []*Column, err error) {
	types, err := t.Migrator().ColumnTypes(tableName)
	if err != nil {
		return nil, err
	}
	for _, column := range types {
		result = append(result, &Column{ColumnType: column, TableName: tableName})
	}
	return result, nil
}

func (t *migrator) GetTableIndex(tableName string) (indexes []gorm.Index, err error) {
	return t.Migrator().GetIndexes(tableName)
}

// =================================================================================

type generator struct {
	config *properties.YamlProperties
}

func (g *generator) ConnectDb() error {
	for _, d := range g.config.Databases {
		u, err := url.Parse(d.DSN)
		if err != nil {
			return err
		}

		db, err := ConnectDB(DBType(u.Scheme), d.DSN)
		if err != nil {
			return err
		}

		d.Db = db
	}

	return nil
}

func (g *generator) GenerateModel(d *properties.Databases) []interface{} {
	ggen := gen.NewGenerator(gen.Config{
		ModelPkgPath: "",
		OutPath:      filepath.Join(d.OutPath, "orm"),
		Mode:         gen.WithDefaultQuery | gen.WithoutContext | gen.WithQueryInterface,

		WithUnitTest: false,

		FieldNullable:     true,
		FieldCoverable:    false,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
		FieldSignable:     false,
	})

	ggen.UseDB(d.Db)

	tableModels := make([]interface{}, 0)

	strTables := strings.Trim(d.StrTables, " ")
	if strTables == "" {
		tableModels = ggen.GenerateAllTable()
	} else {
		for _, t := range strings.Split(strTables, ",") {
			tableModels = append(tableModels, ggen.GenerateModel(strings.Trim(t, " ")))
		}
	}

	if tableModels != nil {
		ggen.ApplyBasic(tableModels...)
		ggen.Execute()
	}

	return tableModels
}

func (g *generator) GenerateDataloader(d *properties.Databases, tableList []interface{}) error {
	ggen := dataloadergen.NewGenerator(dataloadergen.Config{
		OutPath:      filepath.Join(d.OutPath, "dataloader"),
		Package:      "dataloader",
		ModelPackage: path.Join(d.ModuleName, d.OutPath, "model"),
		OrmPackage:   path.Join(d.ModuleName, d.OutPath, "orm"),
	})

	StructFields := make([][]string, 0)
	for _, t := range tableList {
		if t != nil {
		}

		model := &dataloadergen.Model{}
		byteData, _ := json.Marshal(t)
		json.Unmarshal(byteData, &model)

		SFs := ggen.GenerateDataloader(d, *model)
		StructFields = append(StructFields, SFs...)
	}

	ggen.GenerateDataloaderAgg(StructFields)

	return nil
}

func (g *generator) Generate() error {
	err := g.ConnectDb()
	if err != nil {
		return err
	}

	for _, d := range g.config.Databases {
		tableList := g.GenerateModel(d)
		if tableList != nil && d.Dataloader {
			g.GenerateDataloader(d, tableList)
		}
	}

	// tools.Tidy()

	return nil
}

func NewGenerator(config *properties.YamlProperties) *generator {
	return &generator{config}
}

func NewGeneratorFromFile(yaml string) (*generator, error) {
	config, err := LoadConfigFile(yaml)
	if err != nil {
		return nil, err
	}

	return &generator{config}, nil
}

// =================================================================================

type tester struct {
	config *properties.YamlProperties
}

func (g *tester) ConnectDb() error {
	for _, d := range g.config.Databases {
		u, err := url.Parse(d.DSN)
		if err != nil {
			return err
		}

		db, err := ConnectDB(DBType(u.Scheme), d.DSN)
		if err != nil {
			return err
		}

		d.Db = db
	}

	return nil
}

func NewTesterFromFile(yaml string) (*tester, error) {
	config, err := LoadConfigFile(yaml)
	if err != nil {
		return nil, err
	}

	return &tester{config}, nil
}
