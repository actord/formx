package formx

import "fmt"

var globalConfig *Config

func InitGlobalConfig(newFormValidator func() FormValidation) *Config {
	globalConfig = NewConfig(newFormValidator)
	return globalConfig
}

func GetGlobalConfig() *Config {
	if globalConfig == nil {
		panic("formx global config not set")
	}
	return globalConfig
}

type Config struct {
	widgets          map[string]Widget
	widgetError      WidgetError
	newFormValidator func() FormValidation
}

func NewConfig(newFormValidator func() FormValidation) *Config {
	return &Config{
		widgets:          make(map[string]Widget),
		newFormValidator: newFormValidator,
	}
}

func (c *Config) RegisterWidget(widget Widget) error {
	name := widget.GetName()
	if _, ok := c.widgets[name]; ok {
		return fmt.Errorf("widget %s already registered", name)
	}
	c.widgets[name] = widget
	return nil
}

func (c *Config) RegisterWidgetError(widget WidgetError) {
	c.widgetError = widget
}

func (c *Config) RegisterWidgets(widgets ...Widget) error {
	// todo: validate that all default widgets are registered

	for _, widget := range widgets {
		if err := c.RegisterWidget(widget); err != nil {
			return fmt.Errorf("failed to register widget %s: %w", widget.GetName(), err)
		}
	}
	return nil
}

func (c *Config) GetWidget(name string) (Widget, error) {
	widget, ok := c.widgets[name]
	if !ok {
		return nil, fmt.Errorf("widget %s not found", name)
	}
	return widget, nil
}

func GetWidget(name string) (Widget, error) {
	if globalConfig == nil {
		return nil, fmt.Errorf("global config not set")
	}
	return globalConfig.GetWidget(name)
}
