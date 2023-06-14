package api

import (
	"encoding/json"
	"github.com/xhit/go-str2duration/v2"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) (err error) {
	if b[0] == '"' {
		sd := string(b[1 : len(b)-1])
		d.Duration, err = time.ParseDuration(sd)
		return
	}
	var id int64
	id, err = json.Number(string(b)).Int64()
	d.Duration = time.Duration(id)
	return
}
func (d Duration) MarshalJSON() (b []byte, err error) {
	s := str2duration.String(d.Duration)
	return []byte("\"" + s + "\""), nil
}
