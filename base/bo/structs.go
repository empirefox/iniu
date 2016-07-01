// bo split metas into different parts
// each part is use by different logic
package bo

// MetasOutput use to output to client
type MetasOutput struct {
	*StructMetaOut
	Fields []*FieldMetaOut
}

// StructMetaOut use to embed to StructMeta and MetasOutput
type StructMetaOut struct {
	Name      string
	NoDefault bool
	Out       *FormOut
}

// FieldMetaOut use to embed to FieldMeta and MetasOutput
type FieldMetaOut struct {
	Name  string
	Names []string
	Type  string
	Out   *FieldOut
}

// FormOut use to embed to Form and MetasOutput.StructMetaOut
type FormOut struct {
	Detail string `sql:"type:varchar(128)"`
	Pos    int64  `sql:"not null"`
	Title  string `sql:"type:varchar(64)"`
}

// FieldOut use to embed to Field and MetasOutput.FieldMetaOut
type FieldOut struct {
	Detail      string `sql:"type:varchar(128)"`
	Pos         int64  `sql:"not null"`
	Title       string `sql:"type:varchar(64)"`
	Type        string `sql:"type:varchar(16)"`
	Placeholder string `sql:"type:varchar(32)"`
	Help        string `sql:"type:varchar(64)"`
	Ops         string `sql:"type:text"`
}
