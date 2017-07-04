package types

type Upload struct {
  Name     string `redis:"name"     json:"name"`
  Type     string `redis:"type"     json:"type"`
  Category string `redis:"category" json:"category"`
  Path     string `redis:"path"     json:"path"`
}

type ProcessedUpload struct {
  Name     string `redis:"name"     json:"name"`
  Type     string `redis:"type"     json:"type"`
  Category string `redis:"category" json:"category"`
  Path     string `redis:"path"     json:"path"`
  Infohash string `redis:"infohash" json:"infohash"`
}
