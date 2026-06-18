package elements

// import (
// "maps"
// )

type Element struct {
	Type   string            // Type of the element, e.g. "volume", "light", etc.
	Id     string            // Unique identifier for the element, e.g. "speaker1", "lamp2", etc.
	key    string            // key is a unique identifier for the element, e.g. "volume/speaker1"
	config map[string]string // config holds the configuration for the element, e.g. {"min": "0", "max": "100"}
	values map[string]string // values holds the current state of the element, e.g. {"level": "50"}
}

var elements map[string]Element = map[string]Element{}

func NewElement(elementType string, elementId string) Element {
	element := Element{
		Id:     elementId,
		Type:   elementType,
		config: map[string]string{},
		values: map[string]string{},
	}
	elements[elementId] = element
	return element
}

func (e *Element) Init(id string) bool {
	e.Id = id
	return true
}

func (e *Element) GetKey() string {
	return e.Type + "/" + e.Id
}

func (e Element) Set(key, value string) bool {
	return false
}

func (e Element) Loop() bool {
	return false
}

func (e Element) State() map[string]string {
	// var res map[string]string
	// maps.Copy(res, e.config)
	// maps.Copy(res, e.values)
	// return (res)
	return (e.values)
}
