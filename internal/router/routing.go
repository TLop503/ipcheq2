package router

import (
	"io/fs"
	"log"
	"net/http"

	"github.com/tlop503/ipcheq2/v2/internal/web"
)

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
	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
