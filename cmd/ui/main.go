package main

import (
	"context"
	"embed"
	"encoding/json"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/Chinsusu/proxy-server-local/pkg/auth"
	"github.com/Chinsusu/proxy-server-local/pkg/config"
	"log"
)

//go:embed web
var webFS embed.FS

var (
	baseAPI   string
	jwtSecret string
	baseAgent string
)

// PageData is passed to every template render
type PageData struct {
	ActivePage string
	Title      string
}

// renderPage renders base.html + page.html together.
// It parses them fresh per-request so "content" block is always the correct page.
func renderPage(w http.ResponseWriter, page string, data PageData) {
	t, err := template.ParseFS(webFS,
		"web/templates/base.html",
		"web/templates/"+page,
	)
	if err != nil {
		log.Printf("[ERROR] parse template %s: %v", page, err)
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "base", data); err != nil {
		log.Printf("[ERROR] execute template %s: %v", page, err)
	}
}

func main() {
	cfg := config.LoadUI()
	addr := strings.TrimSpace(cfg.Addr)
	if addr == "" {
		addr = ":8081"
	}

	baseAPI = strings.TrimSpace(os.Getenv("PGW_UI_API"))
	if baseAPI == "" {
		baseAPI = "http://127.0.0.1:8080"
	}

	baseAgent = strings.TrimSpace(os.Getenv("PGW_UI_AGENT"))
	if baseAgent == "" {
		baseAgent = "http://127.0.0.1:9090/agent"
	}

	jwtSecret = cfg.JWTSecret

	// Static files from embedded FS
	staticFS, err := fs.Sub(webFS, "web/static")
	if err != nil {
		log.Fatalf("[ERROR] static sub: %v", err)
	}
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Pages
	http.HandleFunc("/", handleRoot)
	http.HandleFunc("/manage", handleManage)
	http.HandleFunc("/proxies", handleProxies)
	http.HandleFunc("/emails", handleEmails)
	http.HandleFunc("/paypal", handlePayPal)
	http.HandleFunc("/income", handleIncome)
	http.HandleFunc("/reports", handleReports)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/logout", handleLogout)

	// API proxy (pass-through to pgw-api)
	http.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, "/api/", baseAPI)
	})

	// Agent proxy (pass-through to pgw-agent)
	http.HandleFunc("/agent/", func(w http.ResponseWriter, r *http.Request) {
		proxyRequest(w, r, "/agent/", baseAgent)
	})

	server := &http.Server{Addr: addr}

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
		sig := <-sigCh
		log.Printf("[INFO] received %s, shutting down...", sig)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(ctx); err != nil {
			log.Printf("[ERROR] shutdown: %v", err)
		}
	}()

	log.Printf("[INFO] pgw-ui listening on %s (API=%s, AGENT=%s)", addr, baseAPI, baseAgent)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("[ERROR] server: %v", err)
	}
}

// --- Page handlers ---

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if !uiAuthorized(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	renderPage(w, "dashboard.html", PageData{ActivePage: "dashboard", Title: "Dashboard"})
}

func handleProxies(w http.ResponseWriter, r *http.Request) {
	if !uiAuthorized(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	renderPage(w, "proxies.html", PageData{ActivePage: "proxies", Title: "Proxy Servers"})
}

func handleManage(w http.ResponseWriter, r *http.Request) {
	if !uiAuthorized(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	renderPage(w, "manage.html", PageData{ActivePage: "manage", Title: "Mappings"})
}

func handleEmails(w http.ResponseWriter, r *http.Request) {
	if !uiAuthorized(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	renderPage(w, "emails.html", PageData{ActivePage: "emails", Title: "Emails"})
}

func handlePayPal(w http.ResponseWriter, r *http.Request) {
	if !uiAuthorized(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	renderPage(w, "paypal.html", PageData{ActivePage: "paypal", Title: "PayPal"})
}

func handleIncome(w http.ResponseWriter, r *http.Request) {
	if !uiAuthorized(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	renderPage(w, "income.html", PageData{ActivePage: "income", Title: "Income"})
}

func handleReports(w http.ResponseWriter, r *http.Request) {
	if !uiAuthorized(r) {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	renderPage(w, "reports.html", PageData{ActivePage: "reports", Title: "Reports"})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		// Render login page (standalone, no base layout)
		t, err := template.ParseFS(webFS, "web/templates/login.html")
		if err != nil {
			http.Error(w, "template error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = t.Execute(w, nil)

	case http.MethodPost:
		var reqBody struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		api := strings.TrimSuffix(baseAPI, "/") + "/v1/auth/login"
		jr, _ := http.NewRequest(http.MethodPost, api, strings.NewReader(string(mustJSON(reqBody))))
		jr.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(jr)
		if err != nil {
			http.Error(w, "upstream unreachable", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			w.WriteHeader(resp.StatusCode)
			io.Copy(w, resp.Body)
			return
		}
		var out struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			http.Error(w, "bad upstream response", http.StatusBadGateway)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:     "pgw_jwt",
			Value:    out.Token,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		w.WriteHeader(http.StatusNoContent)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "pgw_jwt",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	http.Redirect(w, r, "/login", http.StatusFound)
}

// --- Auth ---

func uiAuthorized(r *http.Request) bool {
	c, err := r.Cookie("pgw_jwt")
	if err != nil || c.Value == "" {
		return false
	}
	cl, err := auth.ParseJWT(c.Value, jwtSecret)
	if err != nil {
		return false
	}
	return cl != nil && cl.ExpiresAt != nil && cl.ExpiresAt.Time.After(time.Now())
}

// --- Reverse proxy ---

func proxyRequest(w http.ResponseWriter, r *http.Request, prefix, upstream string) {
	u, err := url.Parse(upstream)
	if err != nil {
		http.Error(w, "invalid upstream", http.StatusInternalServerError)
		return
	}

	path := strings.TrimPrefix(r.URL.Path, prefix)
	targetURL := strings.TrimSuffix(u.String(), "/") + "/" + strings.TrimPrefix(path, "/")
	if r.URL.RawQuery != "" {
		targetURL += "?" + r.URL.RawQuery
	}

	req, err := http.NewRequest(r.Method, targetURL, r.Body)
	if err != nil {
		http.Error(w, "create request failed", http.StatusBadGateway)
		return
	}
	req.Header = r.Header.Clone()
	// Inject JWT from cookie if not in Authorization header
	if req.Header.Get("Authorization") == "" {
		if c, err := r.Cookie("pgw_jwt"); err == nil && c.Value != "" {
			req.Header.Set("Authorization", "Bearer "+c.Value)
		}
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "upstream unreachable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

// --- Helpers ---

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
