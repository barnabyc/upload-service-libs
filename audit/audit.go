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
  AUDIT_EVENTS_KEYSPACE   = "audit:events"
  USERS_ACTIVITY_KEYSPACE = "users:activity"
)

type AuditEvent struct {
  Timestamp  time.Time `json:"timestamp"`
  Type       string    `json:"type"`
  AuditTrail ulid.ULID `json:"audit_trail"`
  Status     string    `json:"status"`
  Bytes      int       `json:"bytes"`
  User       string    `json:"user"`
  Ref        string    `json:"ref"`
}

// todo: import this properly for reuse
type Upload struct {
  Name     string `json:"name"`
  Type     string `json:"type"`
  Category string `json:"category"`
  Path     string `json:"path"`
}

type Activity struct {
  Activity string `json:"activity"`
  Result   string `json:"result"`
  Infohash string `json:"infohash"`
  User     string `json:"user"`
  Key      string `json:"key"`
  Ref      Upload `json:"ref"`
}

func getDateStamp(t time.Time) string {
  foo := fmt.Sprintf("%d%02d%02d", t.Year(), t.Month(), t.Day())

  return foo
}

// func getTimeStamp(t time.Time) string {
//   return fmt.Sprintf("%02d%02d%02d.%d", t.Hour(), t.Minute(), t.Second(), t.Nanosecond())
// }

func buildAuditKey(event AuditEvent) string {
  // group the audit events at day granularity
  stamp := getDateStamp(event.Timestamp)
  components := []string{AUDIT_EVENTS_KEYSPACE, event.Type, stamp}

  return strings.Join(components, ":")
}

func buildActivityKey(activity Activity) string {
  // todo: build this from the Activity
  username := "captainwiggles"
  components := []string{USERS_ACTIVITY_KEYSPACE, username}

  return strings.Join(components, ":")
}

func buildSimpleEvent(thing interface{}, detail string) AuditEvent {
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

func buildActivity(thing interface{}, detail string) Activity {
  // todo: set these values based on provided
  return Activity{
    "upload",
    "created",
    "<no infohash until processed>",
    "captainwiggles",
    detail,
    thing,
  }
}

func newULID(now time.Time) ulid.ULID {
  entropy := rand.New(rand.NewSource(now.UnixNano()))
  return ulid.MustNew(ulid.Timestamp(now), entropy)
}

func Log(conn redis.Conn, thing interface{}, detail string) {
  auditEvent := buildSimpleEvent(thing, detail)
  LogEvent(conn, auditEvent)

  activity := buildActivity(thing, detail)
  LogUserActivity(conn, activity)
}

func LogEvent(conn redis.Conn, event AuditEvent) {
  now := time.Now().UTC()

  trailMarker := newULID(now)

  // store in redis
  _, err := conn.Do(
    "HMSET",
    buildAuditKey(event),
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

func LogUserActivity(conn redis.Conn, activity Activity) {
  json, err := json.Marshal(activity)
  if err != nil {
    log.Printf("Error converting AuditEvent to JSON: %s, %s\n", err, activity)
    return
  }

  _, err := conn.Do(
    "LPUSH",
    buildActivityKey(activity),
    json,
  )
}
