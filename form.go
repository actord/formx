package formx

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"unicode"
)

type FieldInfo struct {
	Name        string
	Label       string
	Placeholder string
	Error       string
	Required    bool
	Tags        map[string]string

	Value reflect.Value
	Type  reflect.Value
}

func GenerateFormInputs(ctx context.Context, formValidation FormValidation, v interface{}) Component {
	rv := valueOf(v)
	if rv.Kind() != reflect.Struct {
		// We can't really do much with a non-struct type. I suppose this
		// could eventually support maps as well, but for now it does not.
		panic("invalid value; only structs are supported")
	}

	t := rv.Type()

	components := make([]Component, 0)

	for i := 0; i < t.NumField(); i++ {
		rf := rv.Field(i)
		fieldName := t.Field(i).Name
		if unicode.IsLower(rune(fieldName[0])) {
			// skip private fields
			continue
		}

		// todo: validate formProps exist here

		tags := parseTags(t.Field(i).Tag.Get("form"))
		//log.Println("tags", tags)

		fieldLabel := t.Field(i).Name
		if label, ok := tags["label"]; ok {
			fieldLabel = label
		}
		fieldRequired := false
		if requiredTag, ok := tags["required"]; ok && requiredTag == "true" {
			fieldRequired = true
		}

		fieldPlaceholder := ""
		if placeholder, ok := tags["placeholder"]; ok {
			fieldPlaceholder = placeholder
		}
		var fieldError string
		if formValidation != nil {
			fieldError = formValidation.ErrorMessage(t.Field(i).Name)
		}

		fieldInfo := FieldInfo{
			Name:        fieldName,
			Label:       fieldLabel,
			Placeholder: fieldPlaceholder,
			Error:       fieldError,
			Required:    fieldRequired,
			Tags:        tags,

			Value: rf,
			Type:  rv,
		}

		widgetName, ok := tags["widget"]
		if !ok {
			switch rf.Kind() {
			case reflect.Struct, reflect.String:
				widgetName = string(DefaultWidgetTypeText)

				if fType, ok := tags["type"]; ok {
					switch fType {
					case "textarea":
						widgetName = string(DefaultWidgetTypeTextarea)
					case "date":
						widgetName = string(DefaultWidgetTypeDate)
					}
				}

			// todo add rf.Kind() == reflect.Int64 default widget for <input type="number">
			case reflect.Bool:
				widgetName = string(DefaultWidgetTypeCheckbox)

			default:
				log.Println("rf:", rf.Kind(), t.Field(i).Name)
				panic("unsupported type")
			}
		}

		widget, err := GetWidget(widgetName)
		if err != nil {
			panic(err)
		}

		component, err := widget.Render(ctx, fieldInfo, formValidation)
		if err != nil {
			panic(err)
		}
		components = append(components, component)
	}

	if formValidation != nil && formValidation.HasFormError() {
		component, err := globalConfig.widgetError.Render(formValidation)
		if err != nil {
			panic(err)
		}
		components = append(components, component)
	}

	return templListComponents{components}
}

func RestoreForm[T any](c *http.Request) (T, FormValidation) {
	form := new(T)
	formValidation := RestoreFormPointer(c, form)
	return *form, formValidation
}

func RestoreFormPointer(r *http.Request, form interface{}) FormValidation {
	formValidation := globalConfig.newFormValidator()

	rv := valueOf(form)
	if rv.Kind() != reflect.Struct {
		// We can't really do much with a non-struct type. I suppose this
		// could eventually support maps as well, but for now it does not.
		panic("invalid value; only structs are supported")
	}

	t := rv.Type()

	for i := 0; i < t.NumField(); i++ {
		rf := rv.Field(i)
		fieldName := t.Field(i).Name
		if unicode.IsLower(rune(fieldName[0])) {
			// skip private fields
			continue
		}

		fieldType := t.Field(i).Type.String()
		formValue := r.FormValue(fieldName)

		//log.Println(fieldName, "(", fieldType, ") = ", formValue)

		if !rf.CanSet() {
			panic(fmt.Sprintf("field %s is not settable", fieldName))
		}

		if formValue == "" {
			// there is empty value, so we can skip setting it
			continue
		}

		switch fieldType {
		case "string":
			rf.SetString(formValue)
		case "int64":
			intVal, err := strconv.ParseInt(formValue, 10, 64)
			if err != nil {
				panic(fmt.Sprintf("field '%s' value '%s' is not int64", fieldName, formValue))
			}
			rf.SetInt(intVal)
		case "bool":
			boolVal := formValue == "on"
			rf.SetBool(boolVal)
		case "formx.Date":
			date := &Date{}
			if err := date.ParseString(formValue); err != nil {
				panic(err)
			}
			rf.Set(reflect.ValueOf(*date))
		case "[]string":
			rf.Set(reflect.ValueOf(r.Form[fieldName]))
		default:
			log.Println(fmt.Sprintf("field '%s' type '%s' is not supported", fieldName, fieldType))
			panic(fmt.Sprintf("field '%s' type '%s' is not supported", fieldName, fieldType))
		}
	}

	formValidation.ValidateStruct(form)

	return formValidation
}

// valueOf is basically just reflect.ValueOf, but if the Kind() of the
// value is a pointer or interface it will try to get the reflect.Value
// of the underlying element, and if the pointer is nil it will
// create a new instance of the type and return the reflect.Value of it.
//
// This is used to make the rest of the fields function simpler.
func valueOf(v interface{}) reflect.Value {
	rv := reflect.ValueOf(v)
	// If a nil pointer is passed in but has a type we can recover, but I
	// really should just panic and tell people to fix their shitty code.
	if rv.Type().Kind() == reflect.Ptr && rv.IsNil() {
		rv = reflect.New(rv.Type().Elem()).Elem()
	}
	// If we have a pointer or interface let's try to get the underlying
	// element
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	return rv
}

func parseTags(tags string) map[string]string {
	tags = strings.TrimSpace(tags)
	if len(tags) == 0 {
		return map[string]string{}
	}
	split := strings.Split(tags, ";")
	ret := make(map[string]string, len(split))
	for _, tag := range split {
		kv := strings.Split(tag, "=")
		if len(kv) < 2 {
			if kv[0] == "-" {
				return map[string]string{
					"-": "this field is ignored",
				}
			}
			continue
		}
		k, v := strings.TrimSpace(kv[0]), strings.TrimSpace(kv[1])
		ret[k] = v
	}
	return ret
}

type templListComponents struct {
	components []Component
}

func (t templListComponents) Render(ctx context.Context, w io.Writer) error {
	for _, component := range t.components {
		err := component.Render(ctx, w)
		if err != nil {
			panic(err)
		}
	}
	return nil
}
