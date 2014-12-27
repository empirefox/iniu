package security

import (
	"reflect"
	"regexp"

	. "github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"

	. "github.com/empirefox/iniu/base"
	. "github.com/empirefox/iniu/gorm/db"
	"github.com/empirefox/shirolet"
)

type (
	WebPerms        map[string]shirolet.Permit
	FormWebPerms    map[Table]WebPerms
	ColumnPerms     map[string]shirolet.Permit
	FormColumnPerms map[Table]ColumnPerms
	JsonFormsMap    map[Table]JsonForm
)

var (
	permNameReg     = regexp.MustCompile(`^([\S]*)Perm$`)
	formWebPerms    = make(FormWebPerms)
	formColumnPerms = make(FormColumnPerms)
)

// exports
var (
	JsonForms  = make(JsonFormsMap)
	WebPerm    = webPerm
	ColumnPerm = columnPerm
	InitForms  = initForms
	InitForm   = initForm
	IdTables   = make(map[int64]string)
)

func webPerm(t Table, method string) shirolet.Permit {
	if webPerms, found := formWebPerms[t]; found {
		return webPerms[method]
	}
	return nil
}

func columnPerm(t Table, column string) shirolet.Permit {
	if columnPerms, found := formColumnPerms[t]; found {
		return columnPerms[column]
	}
	return nil
}

func initForms() {
	forms := []Form{}
	err := DB.Find(&forms).Error
	if err != nil {
		panic(err)
	}

	for _, form := range forms {
		InitForm(form)
	}
}

func initForm(form Form) {
	if form.Id == 0 {
		return
	}

	fields := []Field{}
	err := DB.Set("context:account", "*").Model(form).Related(&fields).Error
	if err != nil && err != gorm.RecordNotFound {
		panic(err)
	}
	form.Fields = fields

	formType := reflect.TypeOf(form)
	formValue := reflect.ValueOf(form)
	tarTable := ToTable(form.Name)
	fieldSize := formType.NumField()

	// webPermMap
	webPermMap := make(WebPerms)
	for i := 0; i < fieldSize; i++ {
		if formWebName := formType.Field(i).Name; permNameReg.MatchString(formWebName) {
			k := permNameReg.ReplaceAllString(formWebName, "$1")
			v := shirolet.NewPermitRaw(formValue.Field(i).String())
			webPermMap[k] = v
		}
	}
	formWebPerms[tarTable] = webPermMap

	// columnPermMap
	columnPermMap := make(ColumnPerms)
	for _, field := range fields {
		formColumnValue := reflect.ValueOf(field)
		k := gorm.ToSnake(formColumnValue.FieldByName("Name").String())
		v := shirolet.NewPermitRaw(formColumnValue.FieldByName("Perm").String())
		columnPermMap[k] = v
	}
	formColumnPerms[tarTable] = columnPermMap

	// Newid to New
	if newId := formValue.FieldByName("Newid").Int(); newId != 0 {
		tarM := New(tarTable)
		if tarM != nil && DB.Table(tarTable).Where("id = ?", newId).First(tarM).Error == nil {
			form.New = tarM
		}
	}

	// Copy
	jform := JsonForm{}
	Copy(&jform, &form)

	// JsonForms
	JsonForms[tarTable] = jform

	// IdTables
	IdTables[form.Id] = tarTable
}
