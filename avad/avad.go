package avad

import (
	"ava/core"
	"github.com/orcaman/concurrent-map"
	"net/http"
	"strings"
	"sync"
)

var ConnStatus = cmap.New()

func Manger(addrs []string) {

	for _, host := range addrs {
		var mutex sync.Mutex
		ConnStatus.Set(host, &core.WsStruct{Status: false, Conn: nil, Mutex: &mutex})
	}

	go ping()

	http.HandleFunc("/exectask", taskRouter)
	http.HandleFunc("/v1/allInfo", getAllInfo)
	http.HandleFunc("/wp", getwp)
	http.HandleFunc("/wpr", getwpr)
	http.HandleFunc("/socks5", socks5Test)
	http.HandleFunc("/v1/proxy", getProxyInfo)
	http.Handle("/", http.FileServer(http.Dir("dist")))

	addr := strings.Join([]string{"0.0.0.0", ":", core.Web}, "")
	http.ListenAndServe(addr, nil)

}
