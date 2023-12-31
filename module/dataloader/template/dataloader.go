package template

const DataloaderPk = `
func Get{{.ModelStructName}}_{{.FieldName}}Loader(Q *orm.Query, redisClient *redis.Client) *{{.ModelStructName}}_{{.FieldName}}Loader {
	return &{{.ModelStructName}}_{{.FieldName}}Loader{
		wait:     2 * time.Millisecond,
		maxBatch: 100,
		fetch: func(keys []{{.Fieldtype}}) ([]*model.{{.ModelStructName}}, []error) {
			resKeys := make([]{{.Fieldtype}}, 0)
			data := make([]*model.{{.ModelStructName}}, len(keys))
			errs := make([]error, len(keys))

			{{if .UseRedis}}
			for i, key := range keys {
				strKey, _ := json.Marshal(key)
				strRec, err := redisClient.Get(context.Background(), fmt.Sprintf("%s_%s_%s", "{{.ModelStructName}}", "{{.FieldName}}", string(strKey))).Result()

				if err != nil {
					resKeys = append(resKeys, key)
					continue
				}

				var rec *model.{{.ModelStructName}}
				err2 := json.Unmarshal([]byte(strRec), &rec)

				if err2 != nil {
					resKeys = append(resKeys, key)
					continue
				}

				data[i] = rec
			}
			{{else}}
			resKeys = keys
			{{end}}

			recs := make([]*model.{{.ModelStructName}}, 0)
			var err error

			if len(resKeys) > 0 {
				recs, err = Q.{{.ModelStructName}}.Where(Q.{{.ModelStructName}}.{{.FieldName}}.In(resKeys...)).Find()
			}

			if err != nil {
				for i := range keys {
					errs[i] = err
				}
				return nil, errs
			}

			for i, key := range keys {
				for _, rec := range recs {
					if key == {{.Asterisk}}rec.{{.FieldName}} {
						data[i] = rec

						{{if .UseRedis}}
						strKey, _ := json.Marshal(key)
						strRec, _ := json.Marshal(rec)
						err := redisClient.Set(context.Background(), fmt.Sprintf("%s_%s_%s", "{{.ModelStructName}}", "{{.FieldName}}", string(strKey)), string(strRec), 30 * time.Second).Err()
						if err != nil {

						}
						{{end}}
					}
				}
			}

			return data, nil
		},
	}
}

`

const DataloaderNpk = `
func Get{{.ModelStructName}}_{{.FieldName}}Loader(Q *orm.Query, redisClient *redis.Client) *{{.ModelStructName}}_{{.FieldName}}Loader {
	return &{{.ModelStructName}}_{{.FieldName}}Loader{
		wait:     2 * time.Millisecond,
		maxBatch: 100,
		fetch: func(keys []{{.Fieldtype}}) ([][]*model.{{.ModelStructName}}, []error) {
			resKeys := make([]{{.Fieldtype}}, 0)
			data := make([][]*model.{{.ModelStructName}}, len(keys))
			errs := make([]error, len(keys))

			{{if .UseRedis}}
			for i, key := range keys {
				strKey, _ := json.Marshal(key)
				strRec, err := redisClient.Get(context.Background(), fmt.Sprintf("%s_%s_%s", "{{.ModelStructName}}", "{{.FieldName}}", string(strKey))).Result()

				if err != nil {
					resKeys = append(resKeys, key)
					continue
				}

				var recs []*model.{{.ModelStructName}}
				err2 := json.Unmarshal([]byte(strRec), &recs)

				if err2 != nil {
					resKeys = append(resKeys, key)
					continue
				}

				data[i] = append(data[i], recs...)
			}
			{{else}}
			resKeys = keys
			{{end}}

			recs := make([]*model.{{.ModelStructName}}, 0)
			var err error

			if len(resKeys) > 0 {
				recs, err = Q.{{.ModelStructName}}.Where(Q.{{.ModelStructName}}.{{.FieldName}}.In(resKeys...)).Find()
			}

			if err != nil {
				for i := range keys {
					errs[i] = err
				}
				return nil, errs
			}

			for i, key := range keys {
				for _, rec := range recs {
					if key == {{.Asterisk}}rec.{{.FieldName}} {
						data[i] = append(data[i], rec)
					}
				}

				{{if .UseRedis}}
				strKey, _ := json.Marshal(key)
				strRec, _ := json.Marshal(data[i])
				err := redisClient.Set(context.Background(), fmt.Sprintf("%s_%s_%s", "Datacatatan", "Tglcatatan", string(strKey)), string(strRec), 30*time.Second).Err()
				if err != nil {

				}
				{{end}}
			}

			return data, nil
		},
	}
}

`

const DataloaderAgg = `
var (
	{{.StrFields}}
)

func SetDefault(Q *orm.Query, redisClient *redis.Client) {
	{{.StrInits}}
}

`
