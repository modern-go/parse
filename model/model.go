package model

// "standard model" to represent parsed object.
// types not in the list below is not part of "standard model".
// it is designed as a intermediate form for convenience instead of performance

// unsigned integer => uint64
// singed integer => int64
// float/double => float64
// text form number => model.Number
// string => string
// []byte => []byte
// bool => bool
// null => null
// map => model.Map
// list/set => model.List

type Number string

type Object interface {
	Get(path ...interface{}) interface{}
}

type Map map[interface{}]interface{}

func (m Map) Get(path ...interface{}) interface{} {
	if len(path) == 0 {
		return m
	}
	if len(path) == 1 {
		return m[path[0]]
	}
	return m[path[0]].(Object).Get(path[1:]...)
}

type List []interface{}

func (l List) Get(path ...interface{}) interface{} {
	if len(path) == 0 {
		return l
	}
	if len(path) == 1 {
		return l[path[0].(int)]
	}
	return l[path[0].(int)].(Object).Get(path[1:]...)
}
