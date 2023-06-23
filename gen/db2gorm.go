package main

import (
	"flag"
	"log"

	v2 "github.com/soedomoto/db2gorm"
)

func main() {
	genPath := flag.String("config", "./db2gorm.yml", "is path for db2gorm.yml")
	flag.Parse()
	if *genPath != "" {
		gen, err := v2.NewGeneratorFromFile(*genPath)
		if err != nil {
			log.Fatalln(err.Error())
		}

		if gen != nil {
			gen.Generate()
			// return
		}
	}

	// gen, _ := v2.NewGeneratorFromFile("./properties.yml")
	// if gen != nil {
	// 	gen.Generate()
	// 	// return
	// }

	// test, _ := NewTesterFromFile("./properties.yml")
	// if test != nil {
	// 	test.ConnectDb()
	// 	redisClient := redis.NewClient(&redis.Options{
	// 		// Addr:     c.Redis.Addr,
	// 		// Username: c.Redis.Username,
	// 		// Password: c.Redis.Password,
	// 		// DB:       Db,
	// 		PoolSize:        10,
	// 		MinIdleConns:    4,
	// 		MaxRetries:      10,
	// 		MinRetryBackoff: 5,
	// 		MaxRetryBackoff: 20,
	// 		ConnMaxIdleTime: 1 * time.Minute,
	// 		// MaxConnAge:   30 * time.Second,
	// 		// IdleTimeout:  cast.ToDuration(redisCfg.IdleTimeout) * time.Millisecond,
	// 	})

	// 	orm.SetDefault(test.config.Databases[0].db)

	// 	dps, err := orm.Q.Datapendidikan.WithContext(context.Background()).Find()
	// 	if err != nil {
	// 		fmt.Errorf(err.Error())
	// 	}

	// 	DatapokokIDs := make([]int64, 0)
	// 	for _, dp := range dps {
	// 		if slices.Contains(DatapokokIDs, *dp.DatapokokID) == false {
	// 			DatapokokIDs = append(DatapokokIDs, *dp.DatapokokID)
	// 		}
	// 	}

	// 	dp2s, err2 := dataloader.GetDatapendidikan_IDLoader(orm.Q, redisClient).LoadAll(DatapokokIDs)
	// 	if err2 != nil && dp2s != nil {
	// 	}

	// 	// cDatapokokIDs := xslice.SplitToChunks(DatapokokIDs, 2000).([][]int64)
	// 	// for _, ids := range cDatapokokIDs {
	// 	// 	dp2s, err2 := orm.Q.Datapokok.WithContext(context.Background()).Where(orm.Datapokok.ID.In(ids...)).Find()
	// 	// 	if err2 != nil && dp2s != nil {
	// 	// 		fmt.Errorf(err2.Error())
	// 	// 	}
	// 	// }

	// }

}
