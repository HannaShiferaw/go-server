// Command go-server is a small web server built with only the Go
// standard library. It serves static pages, answers /hello, and
// accepts a name-and-address form at /form. Submissions are kept
// in memory and can be listed at /submissions.
package main

import (
	"context"
	"fmt"
	"html"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Submission is one form submission.
type Submission struct {
	Name    string
	Address string
}

// store holds submissions in memory, safe for concurrent use.
type store struct {
	mu   sync.Mutex
	list []Submission
}

func (s *store) add(sub Submission) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.list = append(s.list, sub)
}

func (s *store) all() []Submission {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]Submission, len(s.list))
	copy(out, s.list)
	return out
}

var submissions = &store{}

// helloHandler answers GET /hello with a plain greeting.
// Any other method gets 405 Method Not Allowed.
func helloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed. Use GET.", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "Hello!")
}

// formHandler accepts POST /form with name and address fields.
// Both fields are required. The values are escaped before they are
// shown back, so a visitor cannot inject HTML into the page.
func formHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Could not read the form.", http.StatusBadRequest)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	address := strings.TrimSpace(r.FormValue("address"))
	if name == "" || address == "" {
		http.Error(w, "Name and address are both required.", http.StatusBadRequest)
		return
	}

	submissions.add(Submission{Name: name, Address: address})

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html lang="en"><head><meta charset="UTF-8"><title>Thank you</title>
<link rel="stylesheet" href="/style.css"></head><body>
<main><h1>Thank you!</h1>
<p>POST request successful.</p>
<p>Name = %s</p>
<p>Address = %s</p>
<p><a href="/form.html">Submit another</a> &middot; <a href="/submissions">See all submissions</a> &middot; <a href="/">Home</a></p>
</main></body></html>`,
		html.EscapeString(name), html.EscapeString(address))
}

// submissionsHandler lists every submission so far (GET only).
func submissionsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed. Use GET.", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, `<!DOCTYPE html>
<html lang="en"><head><meta charset="UTF-8"><title>Submissions</title>
<link rel="stylesheet" href="/style.css"></head><body>
<main><h1>Submissions</h1>`)
	list := submissions.all()
	if len(list) == 0 {
		fmt.Fprint(w, "<p>No submissions yet.</p>")
	} else {
		fmt.Fprint(w, "<ul>")
		for _, sub := range list {
			fmt.Fprintf(w, "<li>%s &mdash; %s</li>",
				html.EscapeString(sub.Name), html.EscapeString(sub.Address))
		}
		fmt.Fprint(w, "</ul>")
	}
	fmt.Fprint(w, `<p><a href="/form.html">Add a submission</a> &middot; <a href="/">Home</a></p>
</main></body></html>`)
}

// loggingMiddleware logs each request's method, path and duration.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

// port returns the PORT environment variable, or 8080 when it is unset.
func port() string {
	if p := strings.TrimSpace(os.Getenv("PORT")); p != "" {
		return p
	}
	return "8080"
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("./static")))
	mux.HandleFunc("/hello", helloHandler)
	mux.HandleFunc("/form", formHandler)
	mux.HandleFunc("/submissions", submissionsHandler)

	server := &http.Server{
		Addr:         ":" + port(),
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Shut down gracefully on Ctrl+C / SIGTERM instead of dying mid-request.
	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
		<-stop
		log.Println("Shutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("Shutdown error: %v", err)
		}
	}()

	log.Printf("Starting server at http://localhost:%s", port())
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
