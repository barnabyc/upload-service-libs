package upload

import (
  "bytes"
  "crypto/sha1"
  "github.com/garyburd/redigo/redis"
  "github.com/jackpal/bencode-go"
  "log"
  // "github.com/oklog/ulid"
  "reflect"
  "time"
  // "github.com/swatkat/gotrntmetainfoparser" // todo: fork and clean-up
)

// Structs into which torrent metafile is
// parsed and stored into.
type FileDict struct {
  Length int64    "length"
  Path   []string "path"
  Md5sum string   "md5sum"
}

type InfoDict struct {
  FileDuration []int64 "file-duration"
  FileMedia    []int64 "file-media"
  // Single file
  Name   string "name"
  Length int64  "length"
  Md5sum string "md5sum"
  // Multiple files
  Files       []FileDict "files"
  PieceLength int64      "piece length"
  Pieces      string     "pieces"
  Private     int64      "private"
}

type MetaInfo struct {
  Info         InfoDict   "info"
  InfoHash     string     "info hash"
  Announce     string     "announce"
  AnnounceList [][]string "announce-list"
  CreationDate int64      "creation date"
  Comment      string     "comment"
  CreatedBy    string     "created by"
  Encoding     string     "encoding"
}

func Process(ulid []byte, pool *redis.Pool) {
  log.Printf("upload.Process: %s\n", ulid)

  conn := pool.Get()
  defer conn.Close()

  file, err := redis.Bytes(conn.Do("HGET", ulid, "file"))
  log.Printf("debug: file %s\n", reflect.TypeOf(file))

  if err != nil {
    log.Printf("Error getting upload %s\n", err)
    return
  }

  mi := MetaInfo{}

  buf := bytes.NewBuffer(file)

  if mi.ReadTorrentMetaInfo(buf) {
    mi.DumpTorrentMetaInfo()
  } else {
    log.Printf("error: could not parse upload\n")
  }
}

func (metaInfo *MetaInfo) ReadTorrentMetaInfo(buffer *bytes.Buffer) bool {
  // Decode bencoded metainfo file.
  fileMetaData, er := bencode.Decode(buffer)
  if er != nil {
    log.Printf("debug: could not decode file: %s\n", er)
    return false
  }

  // fileMetaData is map of maps of... maps. Get top level map.
  metaInfoMap, ok := fileMetaData.(map[string]interface{})
  if !ok {
    log.Printf("debug: could not get top level map\n")
    return false
  }

  // Enumerate through child maps.
  var bytesBuf bytes.Buffer
  for mapKey, mapVal := range metaInfoMap {
    switch mapKey {
    case "info":
      if er = bencode.Marshal(&bytesBuf, mapVal); er != nil {
        return false
      }

      infoHash := sha1.New()
      infoHash.Write(bytesBuf.Bytes())
      metaInfo.InfoHash = string(infoHash.Sum(nil))

      if er = bencode.Unmarshal(&bytesBuf, &metaInfo.Info); er != nil {
        return false
      }

    case "announce-list":
      if er = bencode.Marshal(&bytesBuf, mapVal); er != nil {
        return false
      }
      if er = bencode.Unmarshal(&bytesBuf, &metaInfo.AnnounceList); er != nil {
        return false
      }

    case "announce":
      metaInfo.Announce = mapVal.(string)

    case "creation date":
      metaInfo.CreationDate = mapVal.(int64)

    case "comment":
      metaInfo.Comment = mapVal.(string)

    case "created by":
      metaInfo.CreatedBy = mapVal.(string)

    case "encoding":
      metaInfo.Encoding = mapVal.(string)
    }
  }

  return true
}

// Print torrent meta info struct data.
func (metaInfo *MetaInfo) DumpTorrentMetaInfo() {
  log.Println("Announce:", metaInfo.Announce)
  log.Println("Announce List:")
  for _, anncListEntry := range metaInfo.AnnounceList {
    for _, elem := range anncListEntry {
      log.Println("    ", elem)
    }
  }
  strCreationDate := time.Unix(metaInfo.CreationDate, 0)
  log.Println("Creation Date:", strCreationDate)
  log.Println("Comment:", metaInfo.Comment)
  log.Println("Created By:", metaInfo.CreatedBy)
  log.Println("Encoding:", metaInfo.Encoding)
  log.Printf("InfoHash: %X\n", metaInfo.InfoHash)
  log.Println("Info:")
  log.Println("    Piece Length:", metaInfo.Info.PieceLength)
  piecesList := metaInfo.getPiecesList()
  log.Printf("    Pieces: %d -- %d\n", len(piecesList), len(metaInfo.Info.Pieces))
  log.Println("    File Duration:", metaInfo.Info.FileDuration)
  log.Println("    File Media:", metaInfo.Info.FileMedia)
  log.Println("    Private:", metaInfo.Info.Private)
  log.Println("    Name:", metaInfo.Info.Name)
  log.Println("    Length:", metaInfo.Info.Length)
  log.Println("    Md5sum:", metaInfo.Info.Md5sum)
  log.Println("    Files:")
  for _, fileDict := range metaInfo.Info.Files {
    log.Println("        Length:", fileDict.Length)
    log.Println("        Path:", fileDict.Path)
    log.Println("        Md5sum:", fileDict.Md5sum)
  }
}

// Splits pieces string into an array of 20 byte SHA1 hashes.
func (metaInfo *MetaInfo) getPiecesList() []string {
  var piecesList []string
  piecesLen := len(metaInfo.Info.Pieces)
  for i, j := 0, 0; i < piecesLen; i, j = i+20, j+1 {
    piecesList = append(piecesList, metaInfo.Info.Pieces[i:i+19])
  }
  return piecesList
}
