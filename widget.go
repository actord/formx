package formx

import (
	"context"
)

type DefaultWidgetType string

const (
	DefaultWidgetTypeText     DefaultWidgetType = "text"
	DefaultWidgetTypeTextarea DefaultWidgetType = "textarea"
	DefaultWidgetTypeDate     DefaultWidgetType = "date"
	DefaultWidgetTypeCheckbox DefaultWidgetType = "checkbox"
)

type Widget interface {
	GetName() string
	Render(ctx context.Context, info FieldInfo, props FormValidation) (Component, error)
}

type WidgetError interface {
	Render(props FormValidation) (Component, error)
}
