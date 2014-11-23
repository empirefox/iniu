package security

import (
	. "github.com/jinzhu/copier"
)

type JsonForm struct {
	Name        string      `json:",omitempty"`
	Description string      `json:",omitempty"`
	Pos         int64       `json:",omitempty"`
	Title       string      `json:",omitempty"`
	JsonFields  []JsonField `json:"Fields,omitempty"`
	New         interface{} `json:",omitempty"`
}

func (jform *JsonForm) Fields(fields []Field) {
	Copy(&jform.JsonFields, &fields)
}

type JsonField struct {
	Name        string `json:",omitempty"`
	Description string `json:",omitempty"`
	Pos         int64  `json:",omitempty"`
	Title       string `json:",omitempty"`
	Type        string `json:",omitempty"`
	Placeholder string `json:",omitempty"`
	Required    bool   `json:",omitempty"`
	Readable    bool   `json:",omitempty"`
	Readonly    bool   `json:",omitempty"`
	Pattern     string `json:",omitempty"`
	Minlength   int    `json:",omitempty"`
	Maxlength   int    `json:",omitempty"`
	Min         string `json:",omitempty"`
	Max         string `json:",omitempty"`
	Help        string `json:",omitempty"`
	Ops         string `json:",omitempty"`
}
