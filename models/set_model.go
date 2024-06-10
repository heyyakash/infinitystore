package models

type Request struct {
	Action string `json:"action"`
	Key    string `json:"key"`
	Value  string `json:"value"`
}
