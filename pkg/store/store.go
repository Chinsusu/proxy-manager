// pkg/store/store.go
package store

import (
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/Chinsusu/proxy-server-local/pkg/types"
)

type Store interface {
	// Proxies
	ListProxies() []types.Proxy
	CreateProxy(p types.Proxy) types.Proxy
	UpdateProxy(p types.Proxy) (types.Proxy, bool)
	DeleteProxy(id string) bool

	// Clients
	ListClients() []types.Client
	CreateClient(c types.Client) types.Client
	DeleteClient(id string) bool

	// Mappings
	ListMappings() []types.MappingView
	CreateMapping(m types.Mapping) (types.MappingView, bool)
	DeleteMapping(id string) bool
	UpdateMappingState(id string, state string, localPort int) bool

	// Telemetry
	SetProxyTelemetry(id string, status types.ProxyStatus, latency int, exitIP string)
	SetProxyTelemetryBatch(updates []TelemetryUpdate)

	// Emails
	ListEmails() []types.Email
	CreateEmail(e types.Email) types.Email
	UpdateEmail(e types.Email) (types.Email, bool)
	DeleteEmail(id string) bool

	// PayPals
	ListPayPals() []types.PayPal
	CreatePayPal(p types.PayPal) types.PayPal
	UpdatePayPal(p types.PayPal) (types.PayPal, bool)
	DeletePayPal(id string) bool

	// Income
	ListIncome() []types.Income
	CreateIncome(i types.Income) types.Income
	DeleteIncome(id string) bool
	GetIncomeReport() types.IncomeReport

	// Nodes
	ListNodes() []types.Node
	CreateNode(n types.Node) types.Node
	GetNode(id string) (types.Node, bool)
	UpdateNode(n types.Node) (types.Node, bool)
	DeleteNode(id string) bool
	UpdateNodeStatus(id, status, version string, lastSeen time.Time) bool
	UpdateNodeDeploy(id, deployStatus, deployLog string) bool
	GetNodeAssignments(nodeID string) types.NodeAssignment
}

// TelemetryUpdate holds a single proxy telemetry result for batch updates.
type TelemetryUpdate struct {
	ID      string
	Status  types.ProxyStatus
	Latency int
	ExitIP  string
}

type memoryStore struct {
	mu       sync.RWMutex
	proxies  map[string]types.Proxy
	clients  map[string]types.Client
	mappings map[string]types.Mapping
	emails   map[string]types.Email
	paypals  map[string]types.PayPal
	income   map[string]types.Income
}

func NewMemory() Store {
	ms := &memoryStore{
		proxies:  make(map[string]types.Proxy),
		clients:  make(map[string]types.Client),
		mappings: make(map[string]types.Mapping),
		emails:   make(map[string]types.Email),
		paypals:  make(map[string]types.PayPal),
		income:   make(map[string]types.Income),
	}

	// Seed demo (có thể bỏ)
	pid := uuid.New().String()
	ms.proxies[pid] = types.Proxy{
		ID:      pid,
		Label:   "demo-http",
		Type:    "http",
		Host:    "127.0.0.1",
		Port:    8088,
		Enabled: true,
		Status:  types.StatusDown,
	}
	cid := uuid.New().String()
	ms.clients[cid] = types.Client{
		ID:      cid,
		IPCidr:  "192.168.2.3/32",
		Enabled: true,
	}
	return ms
}

// ---------- Proxies ----------

func (s *memoryStore) ListProxies() []types.Proxy {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]types.Proxy, 0, len(s.proxies))
	for _, v := range s.proxies { out = append(out, v) }
	return out
}

func (s *memoryStore) CreateProxy(p types.Proxy) types.Proxy {
	s.mu.Lock(); defer s.mu.Unlock()
	if p.ID == "" { p.ID = uuid.New().String() }
	p.Status = types.StatusDown
	s.proxies[p.ID] = p
	return p
}

func (s *memoryStore) UpdateProxy(p types.Proxy) (types.Proxy, bool) {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.proxies[p.ID]; !ok { return types.Proxy{}, false }
	s.proxies[p.ID] = p
	return p, true
}

func (s *memoryStore) DeleteProxy(id string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.proxies[id]; !ok { return false }
	delete(s.proxies, id)
	for mid, m := range s.mappings {
		if m.ProxyID == id { delete(s.mappings, mid) }
	}
	return true
}

// ---------- Clients ----------

func (s *memoryStore) ListClients() []types.Client {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]types.Client, 0, len(s.clients))
	for _, v := range s.clients { out = append(out, v) }
	return out
}

func (s *memoryStore) CreateClient(c types.Client) types.Client {
	s.mu.Lock(); defer s.mu.Unlock()
	if c.ID == "" { c.ID = uuid.New().String() }
	s.clients[c.ID] = c
	return c
}

func (s *memoryStore) DeleteClient(id string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.clients[id]; !ok { return false }
	delete(s.clients, id)
	for mid, m := range s.mappings {
		if m.ClientID == id { delete(s.mappings, mid) }
	}
	return true
}

// ---------- Mappings ----------

func (s *memoryStore) ListMappings() []types.MappingView {
	s.mu.RLock(); defer s.mu.RUnlock()
	type rec struct {
		mv  types.MappingView
		ts  time.Time
		has bool
	}
	tmp := []rec{}
	for _, m := range s.mappings {
		cv, okc := s.clients[m.ClientID]
		pv, okp := s.proxies[m.ProxyID]
		if !okc || !okp { continue }
		r := rec{
			mv: types.MappingView{
				ID:                m.ID,
				Client:            cv,
				Proxy:             pv,
				State:             m.State,
				LocalRedirectPort: m.LocalRedirectPort,
			},
		}
		if m.LastAppliedAt != nil {
			r.ts = *m.LastAppliedAt
			r.has = true
		}
		tmp = append(tmp, r)
	}
	sort.SliceStable(tmp, func(i, j int) bool {
		if tmp[i].has && tmp[j].has {
			return tmp[i].ts.After(tmp[j].ts)
		}
		if tmp[i].has != tmp[j].has {
			return tmp[i].has
		}
		return tmp[i].mv.ID < tmp[j].mv.ID
	})
	out := make([]types.MappingView, len(tmp))
	for i := range tmp { out[i] = tmp[i].mv }
	return out
}

func (s *memoryStore) CreateMapping(m types.Mapping) (types.MappingView, bool) {
	s.mu.Lock(); defer s.mu.Unlock()
	if m.ID == "" { m.ID = uuid.New().String() }
	if _, ok := s.clients[m.ClientID]; !ok { return types.MappingView{}, false }
	if _, ok := s.proxies[m.ProxyID]; !ok { return types.MappingView{}, false }
	m.State = "PENDING"
	s.mappings[m.ID] = m

	cv := s.clients[m.ClientID]
	pv := s.proxies[m.ProxyID]
	return types.MappingView{
		ID:                m.ID,
		Client:            cv,
		Proxy:             pv,
		State:             m.State,
		LocalRedirectPort: m.LocalRedirectPort,
	}, true
}

func (s *memoryStore) DeleteMapping(id string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.mappings[id]; !ok { return false }
	delete(s.mappings, id)
	return true
}

func (s *memoryStore) UpdateMappingState(id string, state string, localPort int) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	m, ok := s.mappings[id]
	if !ok { return false }
	if state != "" {
		m.State = state
		if state == "APPLIED" || state == "FAILED" {
			now := time.Now()
			m.LastAppliedAt = &now
		}
	}
	if localPort > 0 { m.LocalRedirectPort = localPort }
	s.mappings[id] = m
	return true
}

// ---------- Telemetry ----------

func (s *memoryStore) SetProxyTelemetry(id string, status types.ProxyStatus, latency int, exitIP string) {
	s.mu.Lock(); defer s.mu.Unlock()
	s.applyTelemetry(id, status, latency, exitIP)
}

func (s *memoryStore) SetProxyTelemetryBatch(updates []TelemetryUpdate) {
	s.mu.Lock(); defer s.mu.Unlock()
	for _, u := range updates {
		s.applyTelemetry(u.ID, u.Status, u.Latency, u.ExitIP)
	}
}

func (s *memoryStore) applyTelemetry(id string, status types.ProxyStatus, latency int, exitIP string) {
	p, ok := s.proxies[id]; if !ok { return }
	now := time.Now()
	p.Status = status
	if latency > 0 { p.LatencyMs = &latency } else { p.LatencyMs = nil }
	if exitIP != "" { p.ExitIP = &exitIP } else { p.ExitIP = nil }
	p.LastCheckedAt = &now
	s.proxies[id] = p
}

// ---------- Emails ----------

func (s *memoryStore) ListEmails() []types.Email {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]types.Email, 0, len(s.emails))
	for _, v := range s.emails { out = append(out, v) }
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *memoryStore) CreateEmail(e types.Email) types.Email {
	s.mu.Lock(); defer s.mu.Unlock()
	if e.ID == "" { e.ID = uuid.New().String() }
	if e.Status == "" { e.Status = "active" }
	e.CreatedAt = time.Now()
	s.emails[e.ID] = e
	return e
}

func (s *memoryStore) UpdateEmail(e types.Email) (types.Email, bool) {
	s.mu.Lock(); defer s.mu.Unlock()
	old, ok := s.emails[e.ID]
	if !ok { return types.Email{}, false }
	e.CreatedAt = old.CreatedAt // preserve
	s.emails[e.ID] = e
	return e, true
}

func (s *memoryStore) DeleteEmail(id string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.emails[id]; !ok { return false }
	delete(s.emails, id)
	return true
}

// ---------- PayPals ----------

func (s *memoryStore) ListPayPals() []types.PayPal {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]types.PayPal, 0, len(s.paypals))
	for _, v := range s.paypals { out = append(out, v) }
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}

func (s *memoryStore) CreatePayPal(p types.PayPal) types.PayPal {
	s.mu.Lock(); defer s.mu.Unlock()
	if p.ID == "" { p.ID = uuid.New().String() }
	if p.Status == "" { p.Status = "active" }
	if p.Currency == "" { p.Currency = "USD" }
	p.CreatedAt = time.Now()
	s.paypals[p.ID] = p
	return p
}

func (s *memoryStore) UpdatePayPal(p types.PayPal) (types.PayPal, bool) {
	s.mu.Lock(); defer s.mu.Unlock()
	old, ok := s.paypals[p.ID]
	if !ok { return types.PayPal{}, false }
	p.CreatedAt = old.CreatedAt
	s.paypals[p.ID] = p
	return p, true
}

func (s *memoryStore) DeletePayPal(id string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.paypals[id]; !ok { return false }
	delete(s.paypals, id)
	// unlink emails that referenced this paypal
	for eid, e := range s.emails {
		if e.PayPalID != nil && *e.PayPalID == id {
			e.PayPalID = nil
			s.emails[eid] = e
		}
	}
	// cascade: delete income entries linked to this paypal
	for iid, inc := range s.income {
		if inc.PayPalID != nil && *inc.PayPalID == id {
			delete(s.income, iid)
		}
	}
	return true
}

// ---------- Income ----------

func (s *memoryStore) ListIncome() []types.Income {
	s.mu.RLock(); defer s.mu.RUnlock()
	out := make([]types.Income, 0, len(s.income))
	for _, v := range s.income { out = append(out, v) }
	sort.Slice(out, func(i, j int) bool { return out[i].ReceivedAt.After(out[j].ReceivedAt) })
	return out
}

func (s *memoryStore) CreateIncome(i types.Income) types.Income {
	s.mu.Lock(); defer s.mu.Unlock()
	if i.ID == "" { i.ID = uuid.New().String() }
	if i.Currency == "" { i.Currency = "USD" }
	if i.ReceivedAt.IsZero() { i.ReceivedAt = time.Now() }
	i.CreatedAt = time.Now()
	s.income[i.ID] = i
	return i
}

func (s *memoryStore) DeleteIncome(id string) bool {
	s.mu.Lock(); defer s.mu.Unlock()
	if _, ok := s.income[id]; !ok { return false }
	delete(s.income, id)
	return true
}

func (s *memoryStore) GetIncomeReport() types.IncomeReport {
	s.mu.RLock(); defer s.mu.RUnlock()
	report := types.IncomeReport{
		BySource: make(map[string]float64),
		ByMonth:  make(map[string]float64),
	}
	for _, inc := range s.income {
		// Convert to USD (simple 1:1 for now; can add FX rates later)
		amountUSD := inc.Amount
		report.TotalUSD += amountUSD
		report.Count++
		src := inc.Source
		if src == "" { src = "other" }
		report.BySource[src] += amountUSD
		month := inc.ReceivedAt.Format("2006-01")
		report.ByMonth[month] += amountUSD
	}
	return report
}

// ---------- Nodes (memory stub — use sqlite for production) ----------

func (s *memoryStore) ListNodes() []types.Node                                      { return nil }
func (s *memoryStore) CreateNode(n types.Node) types.Node                           { return n }
func (s *memoryStore) GetNode(id string) (types.Node, bool)                         { return types.Node{}, false }
func (s *memoryStore) UpdateNode(n types.Node) (types.Node, bool)                   { return types.Node{}, false }
func (s *memoryStore) DeleteNode(id string) bool                                    { return false }
func (s *memoryStore) UpdateNodeStatus(id, status, version string, lastSeen time.Time) bool { return false }
func (s *memoryStore) UpdateNodeDeploy(id, deployStatus, deployLog string) bool     { return false }
func (s *memoryStore) GetNodeAssignments(nodeID string) types.NodeAssignment {
	return types.NodeAssignment{NodeID: nodeID}
}

