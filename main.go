package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"sync"

	"github.com/go-chi/chi/v5"
)

const (
	server       = "https://ru.wikipedia.org"
	port         = ":8080"
	cacheMaxSize = 20
)

func main() {
	log.SetFlags(log.Lshortfile | log.Ltime)

	r := chi.NewRouter()
	c := newCacheService()

	http.DefaultClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		err := fmt.Errorf("Redirected")
		log.Println("default client check redirect", err.Error())
		return err
	}

	r.Get("/*", func(w http.ResponseWriter, r *http.Request) {

		// 1 берем входящий путь

		u, err := url.Parse(server)
		if err != nil {
			log.Fatal(err)
		}

		u.Path = path.Join(u.Path, r.URL.Path)
		log.Println("path", u.String())

		body, err := c.Get(u.String())
		if err == nil {
			w.Write(body)
			return
		}

		// 2 делаем запрос по входящему путю, ловим редиректы, делаем редирект если нужно
		resp, err := http.Get(u.String())
		if err != nil {
			log.Println("status code", resp.StatusCode)
			log.Println(err.Error())

			if resp.StatusCode == http.StatusMovedPermanently {
				loc, err := resp.Location()
				if err != nil {
					log.Fatal(err)
				}

				http.Redirect(w, r, loc.Path, resp.StatusCode)
				return
			}

			log.Println(resp.StatusCode)
			http.Error(w, err.Error(), 404)
			return
		}

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			http.Error(w, err.Error(), 404)
			return
		}

		c.Add(u.String(), body)
		w.Write(body)
	})

	log.Println("started at :8080")
	log.Println(http.ListenAndServe(":8080", r))
}

type cacheService struct {
	cacheMaxSize int
	sync.Mutex
	cache []cacheEntry
}

type cacheEntry struct {
	url  string
	body []byte
}

func newCacheService() *cacheService {
	return &cacheService{
		cacheMaxSize: cacheMaxSize,
		cache:        make([]cacheEntry, 0, cacheMaxSize),
	}
}

func (c *cacheService) Add(host string, body []byte) {
	c.Lock()
	log.Println("len, cap", len(c.cache), cap(c.cache))
	defer c.Unlock()

	tmpBody := make([]byte, len(body))

	copy(tmpBody, body)

	if len(c.cache) >= c.cacheMaxSize {
		log.Println("add host to oldest entry")

		c.cache[0] = cacheEntry{url: host, body: tmpBody}
		return
	}

	log.Println("add entry")

	c.cache = append(c.cache, cacheEntry{url: host, body: tmpBody})
}

var errEntryNotFound = fmt.Errorf("EntryNotFound")

func (c *cacheService) Get(host string) ([]byte, error) {
	c.Lock()
	defer c.Unlock()

	for _, v := range c.cache {
		if v.url == host {
			bodyCopy := make([]byte, len(v.body))

			copy(bodyCopy, v.body)

			log.Println("use value from cache")

			return bodyCopy, nil
		}
	}

	log.Println("value not found")

	return nil, errEntryNotFound
}
