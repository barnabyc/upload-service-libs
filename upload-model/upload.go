package upload

import (
  "fmt"
  "os"
  "path/filepath"
  "github.com/jackpal/bencode-go"
  "github.com/garyburd/redigo/redis"
  "github.com/swatkat/gotrntmetainfoparser"
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

  mi := gotrntmetainfoparser.MetaInfo{}

  fmt.Printf("debug: upload file ext: %s\n", filepath.Ext(uploadPath))

  file, er := os.Open(uploadPath)
  if er != nil {
    fmt.Printf("debug: could not open file: %s\n", er)
    return
  }
  defer file.Close()

  // Decode bencoded metainfo file.
  fileMetaData, er := bencode.Decode(file)
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





  if mi.ReadTorrentMetaInfoFile( uploadPath ) {
    mi.DumpTorrentMetaInfo()
  } else {
    fmt.Printf("error: could not parse upload\n")
  }
}


