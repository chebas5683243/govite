package govite

import (
	"net/http"
)

func (gv *GoVite) SessionLoad(next http.Handler) http.Handler {
	gv.InfoLog.Println("SessionLoad called")
	return gv.Session.LoadAndSave(next)
}
