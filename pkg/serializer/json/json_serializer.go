package json

import "encoding/json"

type CustomJsonSerializer struct {}

func (c *CustomJsonSerializer) Serialize(v interface{}) ([]byte, error) {
	msg, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
func (c *CustomJsonSerializer) Deserialize(msg []byte, v interface{}) error {
	return json.Unmarshal(msg, v)
}