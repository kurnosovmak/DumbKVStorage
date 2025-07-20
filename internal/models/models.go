package models

type Key struct {
    Key   string `json:"key"`
}

type Value struct {
    Value string `json:"value"`
}

type KeyValue struct {
	Key
	Value
}
