package validate

import (
	"reflect"
	"strings"
	"unicode"
)

type option struct {
	raw    string
	show   string
	negate bool
}

func (opt *option) Err() error {
	raw := opt.raw
	if opt.negate {
		raw = "!" + raw
	}
	return &ValidateErr{Raw: raw, Show: opt.show}
}

func (opt *option) Option() *option {
	return opt
}

func parseOption(opt string) *option {
	negate, showTag := parseNegate(opt)
	tag, show := parseErrShow(showTag)
	return &option{
		raw:    tag,
		show:   show,
		negate: negate,
	}
}

func parseErrShow(tagOpt string) (string, string) {
	ss := strings.Split(tagOpt, "::")
	if len(ss) == 1 {
		return ss[0], ""
	}
	return ss[0], ss[1]
}

// copy from govalidator

// parseNegate Check wether the tag looks like '!something' or 'something'
func parseNegate(tagOpt string) (bool, string) {
	negate := false
	if len(tagOpt) > 0 && tagOpt[0] == '!' {
		tagOpt = string(tagOpt[1:])
		negate = true
	}
	return negate, tagOpt
}

// parseTag splits a struct field's tag into its
// comma-separated options.
func parseTag(typ reflect.Type, tag string) (opts []string) {
	for _, opt := range strings.Split(tag, ",") {
		opt = strings.TrimSpace(opt)
		if ok := isValidTag(opt); ok {
			opts = append(opts, opt)
		} else {
			log.WithField(typ.Name(), opt).Errorln("Invalid validator string")
		}
	}
	return
}

func isValidTag(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		switch {
		case strings.ContainsRune("!#$%&()*+-./:<=>?@[]^_{|}~ ", c):
			// Backslash and quote chars are reserved, but
			// otherwise any punctuation chars are allowed
			// in a tag name.
		default:
			if !unicode.IsLetter(c) && !unicode.IsDigit(c) {
				return false
			}
		}
	}
	return true
}
