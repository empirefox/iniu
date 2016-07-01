// +build go1.5

package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Sirupsen/logrus"
	"github.com/empirefox/pkgs"
	"github.com/jinzhu/gorm"
	"github.com/qor/inflection"
)

var log = logrus.New()

var (
	output = flag.String("output", "", "output file name; default srcdir/<pkg>_inits.go")
)

// Usage is a replacement usage function for the flags package.
func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\tstringer [flags] [directory]\n")
	fmt.Fprintf(os.Stderr, "\tstringer [flags] files... # Must be a single package\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = Usage
	flag.Parse()

	// We accept either one directory or a list of files. Which do we have?
	args := flag.Args()

	pkg := pkgs.NewPackage(args...)
	initer := NewInitFuncStruct(pkg)

	tool, ok := pkg.Tools["iniuiniter"]
	if !ok {
		log.Fatal("iniuiniter options not found")
	}

	for _, typ := range pkg.StructTypes {
		basestruct := pkgs.MustDefaultBool(tool.Types, typ.Name+".basestruct", true)
		if basestruct {
			reg := NewRegisterStruct(typ, pkgs.TryLookupMap(tool.Types, typ.Name+".pkshow"))
			initer.RegisterStructs = append(initer.RegisterStructs, reg)
		}

		tablenames := pkgs.MustDefaultBool(tool.Types, typ.Name+".tablenames", true)
		if tablenames {
			initer.TableNames = append(initer.TableNames, &TableNameStruct{
				Struct: typ.Name,
				Table:  tablename(typ.Name),
			})
		}
	}

	if len(initer.RegisterStructs) > 0 {
		initer.OutputImports["github.com/jinzhu/gorm"] = true
		initer.OutputImports["github.com/empirefox/iniu/base"] = true
	}
	if len(initer.RegisterStructs) > 0 {
		initer.OutputImports["github.com/empirefox/iniu/base"] = true
	}

	src := RenderTemplate(initer)

	// Write to file.
	outputName := *output
	if outputName == "" {
		baseName := fmt.Sprintf("%s_inits.go", pkg.Name)
		outputName = filepath.Join(pkg.Dst, strings.ToLower(baseName))
	}
	os.Rename(outputName, outputName+"~")
	err := ioutil.WriteFile(outputName, src, 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

func RenderTemplate(initer *InitFuncStruct) []byte {
	t := template.Must(template.New("iniuiniter.go").Parse(InitFuncTpl))
	buf := new(bytes.Buffer)
	err := t.Execute(buf, initer)
	if err != nil {
		log.Fatalf("RenderTemplate Execute %s err: %v\n", t.Name(), err)
	}
	src, err := format.Source(buf.Bytes())
	if err != nil {
		log.Fatalf("RenderTemplate format %s err: %v\n%s", t.Name(), err, buf.Bytes())
	}
	return src
}

func NewRegisterStruct(typ *pkgs.Struct, ipkshow interface{}) *RegisterStruct {
	ps := []string{"ID", "Name"}
	switch pkshow := ipkshow.(type) {
	case nil:
	case string:
		ps = strings.Split(pkshow, ",")
	case []string:
		ps = pkshow
	default:
		log.WithField(typ.Name, ipkshow).Fatal("pkshow options invalid")
	}

	if len(ps) != 2 {
		log.WithField(typ.Name, ipkshow).Fatal("pkshow options invalid")
	}

	pks := strings.Split(ps[0], ":")
	shows := strings.Split(ps[1], ":")
	pk := pks[0]
	show := shows[0]
	if pk == "" || show == "" {
		log.WithField(typ.Name, ipkshow).Fatal("pkshow options invalid")
	}

	pkField, ok := typ.IntuitiveFieldMap[pk]
	if !ok {
		log.WithField(typ.Name, ipkshow).Fatal("pkshow options invalid, pk field not found")
	}
	showField, ok := typ.IntuitiveFieldMap[show]
	if !ok {
		log.WithField(typ.Name, ipkshow).Fatal("pkshow options invalid, show field not found")
	}

	p := &RegisterStruct{
		Struct:   typ.Name,
		Pk:       pk,
		Show:     show,
		PkType:   pkField.TypeString,
		ShowType: showField.TypeString,
		Table:    tablename(typ.Name),
		PkDb:     gorm.ToDBName(pk),
		ShowDb:   gorm.ToDBName(show),
	}

	if len(pks) > 1 {
		p.PkDb = pks[1]
	}
	if len(shows) > 1 {
		p.ShowDb = shows[1]
	}

	for name, styp := range typ.Pkg.ArrayTypes {
		if styp.Elem == typ.Name && !styp.IsPtr {
			p.Slice = name
			break
		}
	}
	if p.Slice == "" {
		//		log.WithField("struct", typ.Name).Fatal("has no slice type")
		p.NewSlice = true
		p.Slice = inflection.Plural(typ.Name)
		if _, ok := typ.Pkg.ArrayTypes[p.Slice]; ok {
			p.Slice = typ.Name + "Slice"
		}
	}

	for _, tp := range typ.ComputePkgTagPaths("SA") {
		switch tp.Value {
		case "+":
			p.SearchAttrs = append(p.SearchAttrs, strings.Join(tp.Path, "."))
		case "#":
			if len(tp.Path) == 1 {
				p.SearchAttrs = append(p.SearchAttrs, strings.Join(tp.Path, "."))
			}
		default:
			log.WithField(typ.Name, strings.Join(tp.Path, ".")).Fatal("SA tag must be set +/#")
		}
	}
	return p
}

func tablename(structname string) string {
	return inflection.Plural(gorm.ToDBName(structname))
}

type InitFuncStruct struct {
	PackageName     string
	OutputImports   map[string]bool
	RegisterStructs []*RegisterStruct
	TableNames      []*TableNameStruct
}

func NewInitFuncStruct(pkg *pkgs.Package) *InitFuncStruct {
	return &InitFuncStruct{
		PackageName:   pkg.Name,
		OutputImports: make(map[string]bool),
	}
}

type RegisterStruct struct {
	Struct   string
	Pk       string
	Show     string
	PkType   string
	ShowType string
	Slice    string
	NewSlice bool

	SearchAttrs []string

	Table  string
	PkDb   string
	ShowDb string
}

type TableNameStruct struct {
	Struct string
	Table  string
}

const InitFuncTpl = `
// DO NOT EDIT!
// Code generated by tool

package {{.PackageName}}

import (
{{range $k, $v := .OutputImports}}"{{$k}}"
{{end}}
)

{{range .RegisterStructs}}
	{{if .NewSlice}}
type {{.Slice}} []{{.Struct}}
	{{end}}
{{end}}

func init() {
{{range .RegisterStructs}}
	base.RegisterStruct(
	{{.Struct}}{},
	{{printf "%#v" .SearchAttrs}},

		func() { return new({{.Struct}}) },
		func() { return make({{.Slice}}) },
		func() {
			var s {{.Slice}}
			return s
		},

	[]string{ "{{.Pk}}", "{{.Show}}" },
	func(db *gorm.DB) (results [][]interface{}) {
		ps := []struct {
			{{.Pk}}   {{.PkType}}
			{{.Show}} {{.ShowType}}
		}{}
		db.Table("{{.Table}}").Select("{{.PkDb}}, {{.ShowDb}}").Find(&ps)
		for i := range ps {
			results = append(results, []interface{}{ps[i].{{.Pk}}, ps[i].{{.Show}}})
		}
		return
	})
{{end}}
}

{{range .TableNames}}
func (*{{.Struct}})TableName() string {	return "{{.Table}}" }
{{end}}
`
