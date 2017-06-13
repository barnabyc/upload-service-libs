package upload

import (
  "bytes"
  "fmt"
  "github.com/garyburd/redigo/redis"
  "github.com/jackpal/bencode-go"
  "github.com/oklog/ulid"
  "github.com/swatkat/gotrntmetainfoparser" // todo: fork and clean-up
)

func Process(ulid ULID, pool *redis.Pool) {
  fmt.Printf("upload.Process: %s\n", ulid)

  conn := pool.Get()
  defer conn.Close()

  file, err := redis.String(conn.Do("HGET", ulid, "file"))

  if err != nil {
    fmt.Printf("Error getting upload %s\n", err)
    return
  }

  mi := gotrntmetainfoparser.MetaInfo{}

  buf := bytes.NewBufferString(file)

  // Decode bencoded metainfo file.
  fileMetaData, er := bencode.Decode(buf)
  if er != nil {
    fmt.Printf("debug: could not decode file: %s\n", er)
    return
  }

  // fileMetaData is map of maps of... maps. Get top level map.
  _, ok := fileMetaData.(map[string]interface{})
  if !ok {
    fmt.Printf("debug: could not get top level map\n")
    return
  }

  if mi.ReadTorrentMetaInfoFile(buf) {
    mi.DumpTorrentMetaInfo()
  } else {
    fmt.Printf("error: could not parse upload\n")
  }
}
