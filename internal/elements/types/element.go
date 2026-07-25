package types

type Element struct {
	Type   string
	Id     string
	Config map[string]string
	Values map[string]string
}

func (e *Element) Init(elType string, elId string) bool {
	e.Type = elType
	e.Id = elId
	return true
}

func (e *Element) GetKey() string {
	return e.Type + "/" + e.Id
}

func (e Element) Set(key, value string) bool {
	if e.Config == nil {
		return false
	}
	e.Config[key] = value
	return false
}

func (e Element) Loop() bool {
	return false
}

func (e Element) State() map[string]string {
	return e.Values
}
