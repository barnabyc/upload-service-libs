package audit

import (
  "fmt"
  "github.com/garyburd/redigo/redis"
  "github.com/oklog/ulid"
  "log"
  "math/rand"
  "strings"
  "time"
)

const (
  AUDIT_EVENTS_KEY_PREFIX = "audit:events"
)

type person struct {
  name string
  age  int
}

type AuditEvent struct {
  Timestamp time.Time
  Type      string
  Trail     ulid.ULID
  Status    string
  Bytes     int
  Username  string
  Ref       string
}

func getDateStamp(t time.Time) string {
  foo := fmt.Sprintf("%d%02d%02d", t.Year(), t.Month(), t.Day())

  return foo
}

// func getTimeStamp(t time.Time) string {
//   return fmt.Sprintf("%02d%02d%02d.%d", t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
// }

func buildKey(event AuditEvent) string {
  stamp := getDateStamp(event.Timestamp)
  components := []string{AUDIT_EVENTS_KEY_PREFIX, event.Type, stamp}

  return strings.Join(components, ":")
}

func buildSimpleEvent(thing interface{}, detail string) AuditEvent {
  // Timestamp time.Time
  // Type      string
  // Trail     ulid.ULID
  // Status    string
  // Bytes     int
  // Username  string
  // Ref       string
  now := time.Now().UTC()

  return AuditEvent{
    now,
    "upload", // todo use the interface to get the type
    newULID(now),
    "created",   // todo define statuses
    0,           // todo use upload's size
    "mrgiggles", // todo "
    detail,      // todo "
  }

}

func newULID(now time.Time) ulid.ULID {
  entropy := rand.New(rand.NewSource(now.UnixNano()))
  return ulid.MustNew(ulid.Timestamp(now), entropy)
}

func logTrail() {

}

func LogSimple(conn redis.Conn, thing interface{}, detail string) {
  auditEvent := buildSimpleEvent(thing, detail)
  Log(conn, auditEvent)
}

func Log(conn redis.Conn, event AuditEvent) {
  now := time.Now().UTC()

  trailMarker := newULID(now)

  // store in redis
  _, err := conn.Do(
    "HMSET",
    buildKey(event),
    "trail", trailMarker,
    "status", event.Status,
    "bytes", event.Bytes,
    "username", event.Username,
    "ref", event.Ref,
  )

  if err != nil {
    log.Printf("Error storing audit event %s, %s\n", event, err)
    return
  }

  log.Printf("Audit event logged: %s\n", event)
}
