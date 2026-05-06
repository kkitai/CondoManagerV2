package validator

import (
	"fmt"
	"net/mail"
	"strings"
)

type Errors map[string]string

func (e Errors) Error() string {
	msgs := make([]string, 0, len(e))
	for field, msg := range e {
		msgs = append(msgs, fmt.Sprintf("%s: %s", field, msg))
	}
	return strings.Join(msgs, "; ")
}

func (e Errors) HasErrors() bool {
	return len(e) > 0
}

type Validator struct {
	errors Errors
}

func New() *Validator {
	return &Validator{errors: make(Errors)}
}

func (v *Validator) Required(field, value string) {
	if strings.TrimSpace(value) == "" {
		v.errors[field] = "必須項目です"
	}
}

func (v *Validator) MaxLength(field, value string, max int) {
	if len([]rune(value)) > max {
		v.errors[field] = fmt.Sprintf("%d文字以内で入力してください", max)
	}
}

func (v *Validator) MinLength(field, value string, min int) {
	if len([]rune(value)) < min {
		v.errors[field] = fmt.Sprintf("%d文字以上で入力してください", min)
	}
}

func (v *Validator) Email(field, value string) {
	if value == "" {
		return
	}
	if _, err := mail.ParseAddress(value); err != nil {
		v.errors[field] = "有効なメールアドレスを入力してください"
	}
}

func (v *Validator) OneOf(field, value string, allowed []string) {
	for _, a := range allowed {
		if a == value {
			return
		}
	}
	v.errors[field] = fmt.Sprintf("有効な値を選択してください: %s", strings.Join(allowed, ", "))
}

func (v *Validator) Errors() Errors {
	return v.errors
}

func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}
