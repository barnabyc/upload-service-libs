package audit

import (
	"fmt"
	"github.com/garyburd/redigo/redis"
	"github.com/oklog/ulid"
	"log"
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
	Trail     ULID
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

func buildKey(event AuditEvent, now time.Time) string {
	stamp := getDateStamp(now)
	components := []string{AUDIT_EVENTS_KEY_PREFIX, event.Type, stamp}

	return strings.Join(components, ":")
}

func logEvent() {

}

func logTrail() {

}

func Log(conn redis.Conn, event AuditEvent) {
	now := time.Now().UTC()

	entropy := rand.New(rand.NewSource(now.UnixNano()))
	trailMarker := ulid.MustNew(ulid.Timestamp(now), entropy)

	// store in redis
	_, err := conn.Do(
		"HMSET",
		buildKey(event, now),
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
