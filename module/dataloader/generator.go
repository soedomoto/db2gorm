package dataloader

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	tpl "github.com/soedomoto/db2gorm/module/dataloader/template"

	dataloadgen "github.com/vektah/dataloaden/pkg/generator"
	"golang.org/x/tools/imports"
	"gorm.io/gen/field"
)

func IsDirEmpty(name string) (bool, error) {
	f, err := os.Open(name)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // Or f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err // Either not empty or error, suits both cases
}

func join(sep string, s []string) string {
	return strings.Join(s, sep)
}

func render(tmpl string, wr io.Writer, data interface{}) error {
	t, err := template.New(tmpl).Funcs(template.FuncMap{
		"join": join,
		"getField": func(byref bool, field string) string {
			if byref {
				return "*" + field
			}
			return field
		},
	}).Parse(tmpl)
	if err != nil {
		return err
	}
	return t.Execute(wr, data)
}

func output(fileName string, content []byte) error {
	result, err := imports.Process(fileName, content, nil)
	if err != nil {
		lines := strings.Split(string(content), "\n")
		errLine, _ := strconv.Atoi(strings.Split(err.Error(), ":")[1])
		startLine, endLine := errLine-5, errLine+5
		fmt.Println("Format fail:", errLine, err)
		if startLine < 0 {
			startLine = 0
		}
		if endLine > len(lines)-1 {
			endLine = len(lines) - 1
		}
		for i := startLine; i <= endLine; i++ {
			fmt.Println(i, lines[i])
		}
		return fmt.Errorf("cannot format file: %w", err)
	}
	return ioutil.WriteFile(fileName, result, 0640)
}

func NewGenerator(config Config) *Generator {
	return &Generator{config}
}

type Config struct {
	OutPath      string
	Package      string
	ModelPackage string
	OrmPackage   string
}

type Model struct {
	Generated       bool   // whether to generate db model
	FileName        string // generated file name
	S               string // the first letter(lower case)of simple Name (receiver)
	QueryStructName string // internal query struct name
	ModelStructName string // origin/model struct name
	TableName       string // table name in db server
	// StructInfo      parser.Param
	Fields []*Field
	// Source          model.SourceCode
	ImportPkgPaths []string
	// ModelMethods    []*parser.Method // user custom method bind to db base struct
}

type Field struct {
	Name             string
	Type             string
	ColumnName       string
	ColumnComment    string
	MultilineComment bool
	Tag              field.Tag
	GORMTag          field.GormTag
	CustomGenType    string
	Relation         *field.Relation
}

type Generator struct {
	config Config
}

func (g *Generator) GenerateDataloader(m Model) [][]string {
	StructFields := make([][]string, 0)

	os.MkdirAll(g.config.OutPath, os.ModePerm)
	empty, err := IsDirEmpty(g.config.OutPath)
	if err == nil && empty {
		output(fmt.Sprintf("%s/package.go", g.config.OutPath), []byte(`package dataloader`))
	}

	dataloaderBytes := make([]byte, 0)

	var dataloaderBuf bytes.Buffer
	renderErr := render(tpl.Header, &dataloaderBuf, map[string]interface{}{
		"Package":        g.config.Package,
		"ImportPkgPaths": []string{"github.com/redis/go-redis/v9", g.config.ModelPackage, g.config.OrmPackage},
	})

	if renderErr == nil {
		dataloaderBytes = append(dataloaderBytes, dataloaderBuf.Bytes()...)
	}

	for _, f := range m.Fields {
		Fieldname := f.Name
		Fieldtype := strings.ReplaceAll(f.Type, "*", "")
		Asterisk := ""
		if strings.Contains(f.Type, "*") {
			Asterisk = "*"
		}

		err = dataloadgen.Generate(m.ModelStructName+"_"+Fieldname+"Loader", Fieldtype, "*"+g.config.ModelPackage+"."+m.ModelStructName, g.config.OutPath)
		if err != nil {
			continue
		}

		var dataloaderBuf bytes.Buffer
		renderErr := render(tpl.Dataloader, &dataloaderBuf, map[string]interface{}{
			"Package":         g.config.Package,
			"ImportPkgPaths":  []string{"github.com/redis/go-redis/v9", g.config.ModelPackage, g.config.OrmPackage},
			"ModelStructName": m.ModelStructName,
			"FieldName":       Fieldname,
			"Fieldtype":       Fieldtype,
			"Asterisk":        Asterisk,
		})

		if renderErr == nil {
			dataloaderBytes = append(dataloaderBytes, dataloaderBuf.Bytes()...)
			StructFields = append(StructFields, []string{m.ModelStructName, Fieldname})
		}
	}

	outputErr := output(fmt.Sprintf("%s/%s.gen.go", g.config.OutPath, m.TableName), dataloaderBytes)
	if outputErr == nil {
		return StructFields
	}

	return make([][]string, 0)
}

func (g *Generator) GenerateDataloaderAgg(StructFields [][]string) {
	dataloaderBytes := make([]byte, 0)

	var dataloaderHeaderBuf bytes.Buffer
	renderErr := render(tpl.Header, &dataloaderHeaderBuf, map[string]interface{}{
		"Package":        g.config.Package,
		"ImportPkgPaths": []string{"github.com/redis/go-redis/v9", g.config.ModelPackage, g.config.OrmPackage},
	})

	if renderErr == nil {
		dataloaderBytes = append(dataloaderBytes, dataloaderHeaderBuf.Bytes()...)
	}

	Fields := make([]string, 0)
	Inits := make([]string, 0)
	for _, sf := range StructFields {
		Fields = append(Fields, fmt.Sprintf("%s_%s *%s_%sLoader", sf[0], sf[1], sf[0], sf[1]))
		Inits = append(Inits, fmt.Sprintf("%s_%s= Get%s_%sLoader(Q, redisClient)", sf[0], sf[1], sf[0], sf[1]))
	}

	var dataloaderBuf bytes.Buffer
	renderErr = render(tpl.DataloaderAgg, &dataloaderBuf, map[string]interface{}{
		"Package":        g.config.Package,
		"ImportPkgPaths": []string{"github.com/redis/go-redis/v9", g.config.ModelPackage, g.config.OrmPackage},
		"StrFields":      strings.Join(Fields, "\r\n"),
		"StrInits":       strings.Join(Inits, "\r\n"),
	})

	if renderErr == nil {
		dataloaderBytes = append(dataloaderBytes, dataloaderBuf.Bytes()...)
	}

	outputErr := output(fmt.Sprintf("%s/gen.go", g.config.OutPath), dataloaderBytes)
	if outputErr != nil {

	}
}
