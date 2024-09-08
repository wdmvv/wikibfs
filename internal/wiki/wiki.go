package wiki

import (
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"wikibfs/internal/config"

	"github.com/PuerkitoBio/goquery"
)

// used as a basis for traversing wiki
type Router struct {
	// current wiki instance base, i.e https://en.wikipedia.org
	Base string
	// where to stop and compare
	Finish string

	// home-made goroutines limiter and done
	Lim  chan struct{}
	Done chan struct{}

	// queue with elems to visit
	Elems  chan elem
	Errors chan error

	// map of already visited links
	Were sync.Map
	Wg   sync.WaitGroup
	Mx   sync.Mutex

	// minimal depth so far and respective path
	min     int
	minPath []string

	// stat of how many requests were sent
	Requests atomic.Int32
}

// single element
type elem struct {
	// how did we get here from base
	path  []string
	depth int
}

// both should be /<article>
func Search(base, start, finish string) (int, []string, chan error) {
	r := Router{
		Base:     base,
		Finish:   finish,
		Lim:      make(chan struct{}, config.Config.Goroutines),
		Elems:    make(chan elem),
		Errors:   make(chan error),
		Were:     sync.Map{},
		Wg:       sync.WaitGroup{},
		Mx:       sync.Mutex{},
		min:      0,
		minPath:  make([]string, 0),
		Requests: atomic.Int32{},
	}

	path := make([]string, 0, config.Config.Depth+2)
	path = append(path, start)

	wait := make(chan struct{}, 1)
	go func() {
		wait <- struct{}{}
		r.Elems <- elem{path, 0}
		<-wait
	}()
	wait <- struct{}{}

	r.traverse()
	return r.min, r.minPath, r.Errors
}

func (r *Router) traverse() {
	for {
		select {
		case <-r.Done:
			return
		default:
			r.Wg.Add(1)
			go func() {
				r.Lim <- struct{}{}
				defer func() {
					<-r.Lim
					r.Wg.Done()
				}()

				elem := <-r.Elems

				// if we hit max allowed depth
				// 0 is where we start initially
				if elem.depth == config.Config.Depth {
					return
				}

				// checking last element just in case if it is not an actual endpoint
				// we might be getting /smth#something instead of /something; 35 == #

				last := elem.path[len(elem.path)-1]

				if strings.Contains(last, "#") {
					return
				} else if last == r.Finish {
					r.Mx.Lock()
					if r.min > elem.depth {
						r.min = elem.depth
						r.minPath = elem.path
					}
					r.Mx.Unlock()
					r.Done <- struct{}{}
					return
				}

				// the fun part!!
				if _, ok := r.Were.Load(last); !ok {
					err := r.insertLinks(elem)
					if err != nil {
						r.Errors <- err
					} else {
						r.Were.Store(last, struct{}{})
					}
				}
			}()
		}
	}
}

// gets all hrefs from page and inserts into r.Elems
func (r *Router) insertLinks(e elem) error {
	endpoint := e.path[len(e.path)]
	r.Requests.Add(1)

	resp, err := http.Get(r.Base + endpoint)
	if err != nil || resp.StatusCode != 200 {
		return err
	}

	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	// basically the path to the article itself
	doc.Find("div.mw-page-container").First().Find("div.mw-content-container").Find("main").Find("div.mw-content-ltr").Find("p").Find("p").Find("a").Each(
		func(i int, gq *goquery.Selection) {
			if attr, ok := gq.Attr("href"); ok {
				// in case of recursion
				if attr == endpoint {
					return
				}
				// otherwise we copy path, append and add depth
				var c []string
				copy(e.path, c)
				c = append(c, attr)
				r.Elems <- elem{c, e.depth + 1}
			}
		})

	return nil
}
