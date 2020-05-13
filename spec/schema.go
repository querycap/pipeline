package spec

type Schema struct {
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty" yaml:"type,omitempty"`

	Format string `json:"format,omitempty" yaml:"format,omitempty"`

	Items *Schema `json:"items,omitempty" yaml:"items,omitempty"`

	Required             []string           `json:"required,omitempty" yaml:"required,omitempty"`
	Properties           map[string]*Schema `json:"properties,omitempty" yaml:"properties,omitempty"`
	PropertyNames        *Schema            `json:"propertyNames,omitempty" yaml:"propertyNames,omitempty"`
	AdditionalProperties *Schema            `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`

	AllOf []*Schema `json:"allOf,omitempty"`
	OneOf []*Schema `json:"oneOf,omitempty"`
}
