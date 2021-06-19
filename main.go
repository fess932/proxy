package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"

	"github.com/go-chi/chi/v5"
)

const server = "https://ru.wikipedia.org"

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	r := chi.NewRouter()
	u, err := url.Parse(server)
	if err != nil {
		log.Fatal(err)
	}
	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return fmt.Errorf("Redirected")
	}

	//todo redirect

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		// 1 берем входящий путь
		u.Path = path.Join(u.Path, r.URL.Path)

		log.Println("path", u.String())

		// 2 делаем запрос по входящему путю, ловим редиректы, делаем редирект если нужно
		resp, err := http.Get(u.String())
		if err != nil {
			if resp.StatusCode == http.StatusMovedPermanently {
				loc, err := resp.Location()
				if err != nil {
					log.Fatal(err)
				}

				log.Println("resp loc", loc.Path)
				http.Redirect(w, r, loc.Path, http.StatusMovedPermanently)
				return
			}

			log.Println(resp.StatusCode)
			http.Error(w, err.Error(), 404)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		w.Write(body)
	})

	log.Println("started at :8080")
	log.Println(http.ListenAndServe(":8080", r))
}
