package serializer


type AvalancheSerializer interface {
	Serialize(v interface{}) ([]byte,error)
	Deserialize(msg []byte, v interface{}) error
}