package properties

import (
	"context"
	"log"

	"gorm.io/gorm"
)

type Databases struct {
	Name               string `yaml:"name"`
	DSN                string `yaml:"dsn"` // consult[https://gorm.io/docs/connecting_to_the_database.html]"
	StrTables          string `yaml:"tables"`
	ModuleName         string `yaml:"module_name"`
	OutPath            string `yaml:"out_path"`
	Dataloader         bool   `yaml:"dataloader"`
	DataloaderPkOnly   bool   `yaml:"dataloader_pk_only"`
	DataloaderUseRedis bool   `yaml:"dataloader_use_redis"`
	Db                 *gorm.DB
}

func (db *Databases) info(logInfos ...string) {
	for _, l := range logInfos {
		db.Db.Logger.Info(context.Background(), l)
		log.Println(l)
	}
}

type YamlProperties struct {
	Version   string       `yaml:"version"`   //
	Databases []*Databases `yaml:"databases"` //
}
