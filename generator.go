package v2

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"

	dataloadergen "github.com/soedomoto/db2gorm/module/dataloader"
	"github.com/soedomoto/db2gorm/module/tools"

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

type Databases struct {
	Name string `yaml:"name"`
	DSN  string `yaml:"dsn"` // consult[https://gorm.io/docs/connecting_to_the_database.html]"
	db   *gorm.DB
}

func (db *Databases) info(logInfos ...string) {
	for _, l := range logInfos {
		db.db.Logger.Info(context.Background(), l)
		log.Println(l)
	}
}

type YamlProperties struct {
	Version   string       `yaml:"version"`   //
	Databases []*Databases `yaml:"databases"` //
}

func LoadConfigFile(path string) (*YamlProperties, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close() // nolint
	var props YamlProperties
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
	config *YamlProperties
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

		d.db = db
	}

	return nil
}

func (g *generator) GenerateModel(d *Databases) []interface{} {
	ggen := gen.NewGenerator(gen.Config{
		ModelPkgPath: "",
		OutPath:      "./simpegv2022/orm/",
		Mode:         gen.WithDefaultQuery | gen.WithoutContext | gen.WithQueryInterface,

		WithUnitTest: false,

		FieldNullable:     true,
		FieldCoverable:    false,
		FieldWithIndexTag: true,
		FieldWithTypeTag:  true,
		FieldSignable:     false,
	})

	ggen.UseDB(d.db)

	tableModels := make([]interface{}, 0)
	tableModels = ggen.GenerateAllTable()
	// tableModels = append(tableModels, ggen.GenerateModel("datapokok"))
	// tableModels = append(tableModels, ggen.GenerateModel("datapendidikan"))
	if tableModels != nil {
		ggen.ApplyBasic(tableModels...)
		ggen.Execute()
	}

	return tableModels
}

func (g *generator) GenerateDataloader(tableList []interface{}) error {
	ggen := dataloadergen.NewGenerator(dataloadergen.Config{
		OutPath:      "./simpegv2022/dataloader/",
		Package:      "dataloader",
		ModelPackage: "github.com/soedomoto/db2gorm/simpegv2022/model",
		OrmPackage:   "github.com/soedomoto/db2gorm/simpegv2022/orm",
	})

	for _, t := range tableList {
		if t != nil {
		}

		model := &dataloadergen.Model{}
		byteData, _ := json.Marshal(t)
		json.Unmarshal(byteData, &model)

		ggen.GenerateDataloader(*model)
	}

	return nil
}

func (g *generator) Generate() error {
	err := g.ConnectDb()
	if err != nil {
		return err
	}

	for _, d := range g.config.Databases {
		tableList := g.GenerateModel(d)
		if tableList != nil {
			g.GenerateDataloader(tableList)
		}
	}

	tools.Tidy()

	return nil
}

func NewGenerator(config *YamlProperties) *generator {
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
	config *YamlProperties
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

		d.db = db
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
