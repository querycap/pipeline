package operator

import "github.com/querycap/pipeline/spec"

type WithOperatorMeta interface {
	OperatorMeta() spec.OperatorMeta
}

func OperatorMetaFrom(v interface{}) spec.OperatorMeta {
	if m, ok := v.(WithOperatorMeta); ok {
		return m.OperatorMeta()
	}
	return spec.OperatorMeta{}
}
