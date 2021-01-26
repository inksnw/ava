package avad

import (
	"ava/core"
	"github.com/gorilla/websocket"
	"github.com/orcaman/concurrent-map"
	"net/http"
	"strings"
)

var ConnStatus = cmap.New()

type ConnStruct struct {
	status bool
	conn   *websocket.Conn
}

func Manger(addrs []string) {

	for _, host := range addrs {
		ConnStatus.Set(host, &ConnStruct{false, nil})
	}

	go ping()

	http.HandleFunc("/exectask", taskRouter)
	http.HandleFunc("/v1/allInfo", getAllInfo)
	http.HandleFunc("/v1/proxy", getProxyInfo)
	http.Handle("/", http.FileServer(http.Dir("dist")))

	addr := strings.Join([]string{"0.0.0.0", ":", core.Web}, "")
	http.ListenAndServe(addr, nil)

}
