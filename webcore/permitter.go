package webcore

import (
	"reflect"

	"github.com/empirefox/ffgen/ffgen"
	"github.com/empirefox/iniu/base"
)

func (c *Context) GetPermitter() *permitter {
	if c.permitter == nil {
		c.permitter = c.NewPermitter(c.Resource.Struct)
	}
	return c.permitter
}

func (c *Context) NewPermitter(sm *base.StructMeta) *permitter {
	pf := sm.GetPermittedFields(c.Holds, c.PermType)
	p := &permitter{
		StructMeta:      sm,
		permittedFields: pf,
		permitters:      make(map[reflect.Type]*permitter),
	}
	deps := map[*base.StructMeta]bool{sm: true}
	p.digNextLevel(c, deps)
	return p
}

type permitter struct {
	*base.StructMeta
	permittedFields *base.PermittedFields
	permitters      map[reflect.Type]*permitter
	hasFields       bool
}

func (p *permitter) GetPermitter(typ reflect.Type) (ffgen.Permitter, bool) {
	return p.GetPermitterValidator(typ)
}

func (p *permitter) GetPermitterValidator(typ reflect.Type) (pv ffgen.PermitterValidator, ok bool) {
	if typ == nil {
		return nil, false
	}
	if typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Slice {
		typ = typ.Elem()
	}
	pv, ok = p.permitters[typ]
	return
}

// Below 3 methods implements ffgen.PermitterValidator
func (p *permitter) ValidateField(name string, value interface{}) error {
	return p.Validate(name, value)
}

func (p *permitter) IsPermitted(field string) bool {
	permed, ok := p.permittedFields.Fields[field]
	if ok {
		return permed
	}
	sm, ok := p.permittedFields.StructedFields[field]
	if !ok {
		// default false for not found field
		p.permittedFields.Fields[field] = false
		return false
	}
	fieldPermitter, ok := p.permitters[sm.Mtype]
	if !ok {
		// default false for uncreated permitter
		p.permittedFields.Fields[field] = false
		return false
	}
	ok = fieldPermitter.HasPermitted()
	p.permittedFields.Fields[field] = ok
	return ok
}

func (p *permitter) HasPermitted() bool {
	return p.hasFields
}

func (p *permitter) digNextLevel(c *Context, deps map[*base.StructMeta]bool) {
	if deps[p.StructMeta] {
		return
	}
	deps[p.StructMeta] = true

	for _, fsm := range p.permittedFields.StructedFields {
		if child, ok := p.permitters[fsm.Mtype]; !ok {
			child = &permitter{
				StructMeta:      fsm,
				permittedFields: fsm.GetPermittedFields(c.Holds, c.PermType),
				permitters:      p.permitters,
			}
			p.permitters[fsm.Mtype] = child
			child.digNextLevel(c, deps)
			if len(p.permittedFields.Fields) > 0 || child.hasFields {
				p.hasFields = true
			}
		} else {
			if len(p.permittedFields.Fields) > 0 || child.hasFields {
				p.hasFields = true
			}
		}
	}
}
