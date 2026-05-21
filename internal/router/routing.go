package router

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/tlop503/ipcheq2/v2/internal/web"
)

var Port int

func RouteWebui() {
	// Handle routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		renderTemplate(w, "index.html", Results.Slice())
	})
	http.HandleFunc("/ip", handleIPPost)

	// Serve assets from embedded FS
	assetsFS, _ := fs.Sub(web.FS, "assets")
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.FS(assetsFS))))

	jsFS, _ := fs.Sub(web.FS, "js")
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.FS(jsFS))))
}

func RouteAPI() {
	http.HandleFunc("/api/firstparty", handleFirstPartyGet)
	http.HandleFunc("/api/thirdparty", handleThirdPartyGet)
	http.HandleFunc("/api/fullquery", handleFullQuery)
}

func StartServing() {
	log.Printf("Starting server on :%d\n", Port)
	addr := fmt.Sprintf(":%d", Port)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
