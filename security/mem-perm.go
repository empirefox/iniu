package security

import (
	"reflect"
	"regexp"

	. "github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"

	. "github.com/empirefox/iniu/gorm/db"
	"github.com/empirefox/shirolet"
)

type (
	WebPerms       map[string]shirolet.Permit
	FormWebPerms   map[Table]WebPerms
	FieldPerms     map[string]shirolet.Permit
	FormFieldPerms map[Table]FieldPerms
	JsonFormsMap   map[Table]JsonForm
)

var (
	permNameReg = regexp.MustCompile(`^([\S]*)Perm$`)
)

func init() {
	formWebPerms = make(FormWebPerms)
	formFieldPerms = make(FormFieldPerms)
	JsonForms = make(JsonFormsMap)
}

func WebPerm(t Table, method string) shirolet.Permit {
	if webPerms, found := formWebPerms[t]; found {
		return webPerms[method]
	}
	return nil
}

func FieldPerm(t Table, field string) shirolet.Permit {
	if fieldPerms, found := formFieldPerms[t]; found {
		return fieldPerms[field]
	}
	return nil
}

func InitForms() {
	forms := []Form{}
	err := DB.Find(&forms).Error
	if err != nil {
		panic(err)
	}

	for _, form := range forms {
		InitForm(&form)
	}
}

func InitForm(formPtr *Form) {
	fields := []Field{}
	err := DB.Model(formPtr).Related(&fields).Error
	if err != nil && err != gorm.RecordNotFound {
		panic(err)
	}
	formPtr.Fields = fields

	formType := reflect.TypeOf(*formPtr)
	formValue := reflect.ValueOf(*formPtr)
	tarTable := ToTable(formValue.FieldByName("Name").String())
	fieldSize := formType.NumField()

	// webPermMap
	webPermMap := make(WebPerms)
	for i := 0; i < fieldSize; i++ {
		if formWebName := formType.Field(i).Name; permNameReg.MatchString(formWebName) {
			k := permNameReg.ReplaceAllString(formWebName, "$1")
			v := shirolet.NewPermitRaw(formValue.Field(i).String())
			formWebPermMap[k] = v
		}
	}
	formWebPerms[tarTable] = webPermMap

	// fieldPermMap
	fieldPermMap := make(FieldPerms)
	for _, field := range fields {
		formFieldValue := reflect.ValueOf(field)
		k := formFieldValue.FieldByName("Name").String()
		v := shirolet.NewPermitRaw(formFieldValue.FieldByName("Perm").String())
		fieldPermMap[k] = v
	}
	formFieldPerms[tarTable] = fieldPermMap

	// Newid to New
	if newId := formValue.FieldByName("Newid").Int(); newId != 0 {
		tarM := New(tarTable)
		if tarM != nil && DB.Table(tarTable).Where("id = ?", newId).First(tarM).Error == nil {
			formPtr.New = tarM
		}
	}

	// Copy
	jform := &JsonForm{}
	Copy(formPtr, jform)

	// JsonForms
	JsonForms[tarTable] = jform
}
