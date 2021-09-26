package setting

import (
	"time"
)

var Server = server{}

type server struct {
	Port string
}

var Session = session{}

type session struct {
	CookieName   string
	CookieExpire time.Duration
}

func Load() {
	Server.Port = ":3000"
	Session.CookieName = "gowebserver_session_id"
	Session.CookieExpire = (1 * time.Hour)
}
