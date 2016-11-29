package upload

import (
  "fmt"
  "log"
  "github.com/garyburd/redigo/redis"
  // "github.com/swatkat/gotrntmetainfoparser"
)

func Process(uuid []byte, pool *redis.Pool) {
  fmt.Printf("upload.Process: %s\n", uuid)

  conn := pool.Get()
  defer conn.Close()

  uploadPath, err := redis.String(conn.Do("HGET", string(uuid), "path"))

  if err != nil {
    fmt.Printf("Error getting upload detail %s\n", err)
    return
  }

  fmt.Printf("upload.Path: %s\n", uploadPath)

  // MetaInfo::ReadTorrentMetaInfoFile( uploadPath )
  // MetaInfo::DumpTorrentMetaInfo()

  fmt.Printf("todo: process the upload\n")
  // todo: parse newly upload torrent
}


