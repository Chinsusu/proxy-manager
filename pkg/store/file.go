// pkg/store/file.go
package store

import (
	"crypto/sha256"
	"encoding/json"
	"sort"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/Chinsusu/proxy-server-local/pkg/types"
)

type fileState struct {
	Proxies  map[string]types.Proxy   `json:"proxies"`
	Clients  map[string]types.Client  `json:"clients"`
	Mappings map[string]types.Mapping `json:"mappings"`
	Emails   map[string]types.Email   `json:"emails"`
	PayPals  map[string]types.PayPal  `json:"paypals"`
	Income   map[string]types.Income  `json:"income"`
}

type fileStore struct {
	mu           sync.RWMutex
	path         string
	state        fileState
	lastSaveHash [32]byte
}

func NewFile(path string) Store {
	_ = os.MkdirAll(filepath.Dir(path), 0o750)
	fs := &fileStore{path: path}
	if err := fs.load(); err != nil {
		fs.state = fileState{
			Proxies:  map[string]types.Proxy{},
			Clients:  map[string]types.Client{},
			Mappings: map[string]types.Mapping{},
			Emails:   map[string]types.Email{},
			PayPals:  map[string]types.PayPal{},
			Income:   map[string]types.Income{},
		}
		_ = fs.save()
	}
	return fs
}

func (s *fileStore) load() error {
	b, err := os.ReadFile(s.path)
	if err != nil { return err }
	var st fileState
	if err := json.Unmarshal(b, &st); err != nil { return err }
	if st.Proxies == nil  { st.Proxies  = map[string]types.Proxy{} }
	if st.Clients == nil  { st.Clients  = map[string]types.Client{} }
	if st.Mappings == nil { st.Mappings = map[string]types.Mapping{} }
	if st.Emails == nil   { st.Emails   = map[string]types.Email{} }
	if st.PayPals == nil  { st.PayPals  = map[string]types.PayPal{} }
	if st.Income == nil   { st.Income   = map[string]types.Income{} }
	s.state = st
	s.lastSaveHash = sha256.Sum256(b)
	return nil
}

func (s *fileStore) save() error {
	b, err := json.MarshalIndent(s.state, "", "  ")
	if err != nil { return err }
	h := sha256.Sum256(b)
	if h == s.lastSaveHash { return nil }
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o640); err != nil { return err }
	if err := os.Rename(tmp, s.path); err != nil { return err }
	s.lastSaveHash = h
	return nil
}

// ---------- Proxies ----------

func (s *fileStore) ListProxies() []types.Proxy {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]types.Proxy, 0, len(s.state.Proxies))
	for _, v := range s.state.Proxies { out = append(out, v) }
	return out
}

func (s *fileStore) CreateProxy(p types.Proxy) types.Proxy {
	s.mu.Lock(); defer s.mu.Unlock()
	if p.ID == "" { p.ID = uuid.New().String() }
	p.Status = types.StatusDown
	if s.state.Proxies == nil { s.state.Proxies = map[string]types.Proxy{} }
	s.state.Proxies[p.ID] = p
	_ = s.save()
	return p
}

func (s *fileStore) UpdateProxy(p types.Proxy) (types.Proxy, bool) {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.state.Proxies[p.ID]; !ok { return types.Proxy{}, false }
	s.state.Proxies[p.ID] = p
	_ = s.save()
	return p, true
}

func (s *fileStore) DeleteProxy(id string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.state.Proxies[id]; !ok { return false }
	delete(s.state.Proxies, id)
	for mid, m := range s.state.Mappings {
		if m.ProxyID == id { delete(s.state.Mappings, mid) }
	}
	_ = s.save()
	return true
}

// ---------- Clients ----------

func (s *fileStore) ListClients() []types.Client {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]types.Client, 0, len(s.state.Clients))
	for _, v := range s.state.Clients { out = append(out, v) }
	return out
}

func (s *fileStore) CreateClient(c types.Client) types.Client {
	s.mu.Lock(); defer s.mu.Unlock()
	if c.ID == "" { c.ID = uuid.New().String() }
	if s.state.Clients == nil { s.state.Clients = map[string]types.Client{} }
	s.state.Clients[c.ID] = c
	_ = s.save()
	return c
}

func (s *fileStore) DeleteClient(id string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.state.Clients[id]; !ok { return false }
	delete(s.state.Clients, id)
	for mid, m := range s.state.Mappings {
		if m.ClientID == id { delete(s.state.Mappings, mid) }
	}
	_ = s.save()
	return true
}

// ---------- Mappings ----------

func (s *fileStore) ListMappings() []types.MappingView {
	s.mu.RLock(); defer s.mu.RUnlock()
	type rec struct{
		mv types.MappingView
		ts time.Time
		has bool
	}
	tmp := []rec{}
	for _, m := range s.state.Mappings {
		cv, okc := s.state.Clients[m.ClientID]
		pv, okp := s.state.Proxies[m.ProxyID]
		if !okc || !okp { continue }
		r := rec{ mv: types.MappingView{
			ID: m.ID,
			Client: cv,
			Proxy: pv,
			State: m.State,
			LocalRedirectPort: m.LocalRedirectPort,
		}}
		if m.LastAppliedAt != nil { r.ts = *m.LastAppliedAt; r.has = true }
		tmp = append(tmp, r)
	}
	sort.SliceStable(tmp, func(i,j int) bool {
		if tmp[i].has && tmp[j].has { return tmp[i].ts.After(tmp[j].ts) }
		if tmp[i].has != tmp[j].has { return tmp[i].has }
		return tmp[i].mv.ID < tmp[j].mv.ID
	})
	out := make([]types.MappingView, len(tmp))
	for i := range tmp { out[i] = tmp[i].mv }
	return out
}

func (s *fileStore) CreateMapping(m types.Mapping) (types.MappingView, bool) {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.state.Clients[m.ClientID]; !ok { return types.MappingView{}, false }
	if _, ok := s.state.Proxies[m.ProxyID]; !ok { return types.MappingView{}, false }
	if m.ID == "" { m.ID = uuid.New().String() }
	m.State = "PENDING"
	if s.state.Mappings == nil { s.state.Mappings = map[string]types.Mapping{} }
	s.state.Mappings[m.ID] = m
	_ = s.save()
	cv := s.state.Clients[m.ClientID]
	pv := s.state.Proxies[m.ProxyID]
	return types.MappingView{
		ID: m.ID, Client: cv, Proxy: pv, State: m.State, LocalRedirectPort: m.LocalRedirectPort,
	}, true
}

func (s *fileStore) DeleteMapping(id string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.state.Mappings[id]; !ok { return false }
	delete(s.state.Mappings, id)
	_ = s.save()
	return true
}

func (s *fileStore) UpdateMappingState(id string, state string, localPort int) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	m, ok := s.state.Mappings[id]
	if !ok { return false }
	if state != "" {
		m.State = state
		if state == "APPLIED" || state == "FAILED" {
			now := time.Now()
			m.LastAppliedAt = &now
		}
	}
	if localPort > 0 { m.LocalRedirectPort = localPort }
	s.state.Mappings[id] = m
	_ = s.save()
	return true
}

func (s *fileStore) UpdateMappingNode(id, nodeID string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	m, ok := s.state.Mappings[id]
	if !ok { return false }
	if nodeID == "" {
		m.NodeID = nil
	} else {
		m.NodeID = &nodeID
	}
	s.state.Mappings[id] = m
	_ = s.save()
	return true
}

// ---------- Telemetry ----------

func (s *fileStore) SetProxyTelemetry(id string, status types.ProxyStatus, latency int, exitIP string) {
	s.mu.Lock(); defer s.mu.Unlock()
	s.applyTelemetry(id, status, latency, exitIP)
	_ = s.save()
}

func (s *fileStore) SetProxyTelemetryBatch(updates []TelemetryUpdate) {
	s.mu.Lock(); defer s.mu.Unlock()
	for _, u := range updates {
		s.applyTelemetry(u.ID, u.Status, u.Latency, u.ExitIP)
	}
	_ = s.save()
}

func (s *fileStore) applyTelemetry(id string, status types.ProxyStatus, latency int, exitIP string) {
	p, ok := s.state.Proxies[id]; if !ok { return }
	now := time.Now()
	p.Status = status
	if latency > 0 { p.LatencyMs = &latency } else { p.LatencyMs = nil }
	if exitIP != "" { p.ExitIP = &exitIP } else { p.ExitIP = nil }
	p.LastCheckedAt = &now
	s.state.Proxies[id] = p
}

// ---------- Emails ----------

func (s *fileStore) ListEmails() []types.Email {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]types.Email, 0, len(s.state.Emails))
	for _, v := range s.state.Emails { out = append(out, v) }
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *fileStore) CreateEmail(e types.Email) types.Email {
	s.mu.Lock(); defer s.mu.Unlock()
	if e.ID == "" { e.ID = uuid.New().String() }
	if e.Status == "" { e.Status = "active" }
	e.CreatedAt = time.Now()
	if s.state.Emails == nil { s.state.Emails = map[string]types.Email{} }
	s.state.Emails[e.ID] = e
	_ = s.save()
	return e
}

func (s *fileStore) UpdateEmail(e types.Email) (types.Email, bool) {
	s.mu.Lock(); defer s.mu.Unlock()
	old, ok := s.state.Emails[e.ID]
	if !ok { return types.Email{}, false }
	e.CreatedAt = old.CreatedAt
	s.state.Emails[e.ID] = e
	_ = s.save()
	return e, true
}

func (s *fileStore) DeleteEmail(id string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.state.Emails[id]; !ok { return false }
	delete(s.state.Emails, id)
	_ = s.save()
	return true
}

// ---------- PayPals ----------

func (s *fileStore) ListPayPals() []types.PayPal {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]types.PayPal, 0, len(s.state.PayPals))
	for _, v := range s.state.PayPals { out = append(out, v) }
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *fileStore) CreatePayPal(p types.PayPal) types.PayPal {
	s.mu.Lock(); defer s.mu.Unlock()
	if p.ID == "" { p.ID = uuid.New().String() }
	if p.Status == "" { p.Status = "active" }
	if p.Currency == "" { p.Currency = "USD" }
	p.CreatedAt = time.Now()
	if s.state.PayPals == nil { s.state.PayPals = map[string]types.PayPal{} }
	s.state.PayPals[p.ID] = p
	_ = s.save()
	return p
}

func (s *fileStore) UpdatePayPal(p types.PayPal) (types.PayPal, bool) {
	s.mu.Lock(); defer s.mu.Unlock()
	old, ok := s.state.PayPals[p.ID]
	if !ok { return types.PayPal{}, false }
	p.CreatedAt = old.CreatedAt
	s.state.PayPals[p.ID] = p
	_ = s.save()
	return p, true
}

func (s *fileStore) DeletePayPal(id string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.state.PayPals[id]; !ok { return false }
	delete(s.state.PayPals, id)
	// unlink emails
	for eid, e := range s.state.Emails {
		if e.PayPalID != nil && *e.PayPalID == id {
			e.PayPalID = nil
			s.state.Emails[eid] = e
		}
	}
	// cascade: delete income linked to this paypal
	for iid, inc := range s.state.Income {
		if inc.PayPalID != nil && *inc.PayPalID == id {
			delete(s.state.Income, iid)
		}
	}
	_ = s.save()
	return true
}

// ---------- Income ----------

func (s *fileStore) ListIncome() []types.Income {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]types.Income, 0, len(s.state.Income))
	for _, v := range s.state.Income { out = append(out, v) }
	sort.Slice(out, func(i, j int) bool { return out[i].ReceivedAt.After(out[j].ReceivedAt) })
	return out
}

func (s *fileStore) CreateIncome(i types.Income) types.Income {
	s.mu.Lock(); defer s.mu.Unlock()
	if i.ID == "" { i.ID = uuid.New().String() }
	if i.Currency == "" { i.Currency = "USD" }
	if i.ReceivedAt.IsZero() { i.ReceivedAt = time.Now() }
	i.CreatedAt = time.Now()
	if s.state.Income == nil { s.state.Income = map[string]types.Income{} }
	s.state.Income[i.ID] = i
	_ = s.save()
	return i
}

func (s *fileStore) DeleteIncome(id string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.state.Income[id]; !ok { return false }
	delete(s.state.Income, id)
	_ = s.save()
	return true
}

func (s *fileStore) GetIncomeReport() types.IncomeReport {
	s.mu.RLock(); defer s.mu.RUnlock()
	report := types.IncomeReport{
		BySource: make(map[string]float64),
		ByMonth:  make(map[string]float64),
	}
	for _, inc := range s.state.Income {
		report.TotalUSD += inc.Amount
		report.Count++
		src := inc.Source
		if src == "" { src = "other" }
		report.BySource[src] += inc.Amount
		month := inc.ReceivedAt.Format("2006-01")
		report.ByMonth[month] += inc.Amount
	}
	return report
}

// ---------- Nodes (file store stub — node management requires sqlite) ----------

func (s *fileStore) ListNodes() []types.Node                                           { return nil }
func (s *fileStore) CreateNode(n types.Node) types.Node                                { return n }
func (s *fileStore) GetNode(id string) (types.Node, bool)                              { return types.Node{}, false }
func (s *fileStore) UpdateNode(n types.Node) (types.Node, bool)                        { return types.Node{}, false }
func (s *fileStore) DeleteNode(id string) bool                                         { return false }
func (s *fileStore) UpdateNodeStatus(id, status, version string, lastSeen time.Time) bool { return false }
func (s *fileStore) UpdateNodeDeploy(id, deployStatus, deployLog string) bool          { return false }
func (s *fileStore) GetNodeAssignments(nodeID string) types.NodeAssignment {
	return types.NodeAssignment{NodeID: nodeID}
}

