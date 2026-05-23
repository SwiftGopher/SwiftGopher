package tests

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"

	"swift-gopher/internal/usecase"
	"swift-gopher/pkg/modules"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

// ── Auth ─────────────────────────────────────────────────────────────────────

type mockAuthRepo struct {
	byEmail map[string]*modules.User
	byID    map[string]*modules.User
}

func newMockAuthRepo() *mockAuthRepo {
	return &mockAuthRepo{
		byEmail: make(map[string]*modules.User),
		byID:    make(map[string]*modules.User),
	}
}

func (m *mockAuthRepo) CreateUser(_ context.Context, u *modules.User) error {
	if _, exists := m.byEmail[u.Email]; exists {
		return errors.New("duplicate email")
	}
	m.byEmail[u.Email] = u
	m.byID[u.ID] = u
	return nil
}

func (m *mockAuthRepo) GetUserByEmail(_ context.Context, email string) (*modules.User, error) {
	u, ok := m.byEmail[email]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

func (m *mockAuthRepo) GetUserByID(_ context.Context, id string) (*modules.User, error) {
	u, ok := m.byID[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	return u, nil
}

// ── Courier ───────────────────────────────────────────────────────────────────

type mockCourierRepo struct {
	couriers map[string]*modules.Courier
}

func newMockCourierRepo() *mockCourierRepo {
	return &mockCourierRepo{couriers: make(map[string]*modules.Courier)}
}

func (m *mockCourierRepo) Create(_ context.Context, c *modules.Courier) error {
	m.couriers[c.ID] = c
	return nil
}

func (m *mockCourierRepo) GetByID(_ context.Context, id string) (*modules.Courier, error) {
	c, ok := m.couriers[id]
	if !ok {
		return nil, errors.New("courier not found")
	}
	return c, nil
}

func (m *mockCourierRepo) GetByUserID(_ context.Context, userID string) (*modules.Courier, error) {
	for _, c := range m.couriers {
		if c.UserID == userID {
			return c, nil
		}
	}
	return nil, errors.New("courier not found")
}

func (m *mockCourierRepo) List(_ context.Context) ([]*modules.Courier, error) {
	list := make([]*modules.Courier, 0, len(m.couriers))
	for _, c := range m.couriers {
		list = append(list, c)
	}
	return list, nil
}

func (m *mockCourierRepo) ListFree(_ context.Context) ([]*modules.Courier, error) {
	var free []*modules.Courier
	for _, c := range m.couriers {
		if c.Status == modules.StatusFree {
			free = append(free, c)
		}
	}
	return free, nil
}

func (m *mockCourierRepo) UpdateStatus(_ context.Context, id string, status modules.CourierStatus) error {
	c, ok := m.couriers[id]
	if !ok {
		return errors.New("courier not found")
	}
	c.Status = status
	return nil
}

func (m *mockCourierRepo) UpdateLocation(_ context.Context, id string, lat, lng float64) error {
	c, ok := m.couriers[id]
	if !ok {
		return errors.New("courier not found")
	}
	c.CurrentLat = lat
	c.CurrentLng = lng
	return nil
}

func (m *mockCourierRepo) UpdateTransport(_ context.Context, id string, transport modules.TransportType) error {
	c, ok := m.couriers[id]
	if !ok {
		return errors.New("courier not found")
	}
	c.TransportType = transport
	return nil
}

// ── Order ─────────────────────────────────────────────────────────────────────

type mockOrderRepo struct {
	mu      sync.Mutex
	orders  map[string]*modules.Order
	history []*modules.OrderHistory
}

func newMockOrderRepo() *mockOrderRepo {
	return &mockOrderRepo{orders: make(map[string]*modules.Order)}
}

func (m *mockOrderRepo) Create(_ context.Context, o *modules.Order) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	cp := *o
	m.orders[o.ID] = &cp
	return nil
}

func (m *mockOrderRepo) GetByID(_ context.Context, id string) (*modules.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return nil, usecase.ErrOrderNotFound
	}
	cp := *o
	return &cp, nil
}

func (m *mockOrderRepo) List(_ context.Context) ([]*modules.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]*modules.Order, 0, len(m.orders))
	for _, o := range m.orders {
		cp := *o
		out = append(out, &cp)
	}
	return out, nil
}

func (m *mockOrderRepo) ListByStatus(_ context.Context, status modules.OrderStatus) ([]*modules.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*modules.Order
	for _, o := range m.orders {
		if o.Status == status {
			cp := *o
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *mockOrderRepo) ListWithFilter(_ context.Context, filter modules.OrderFilter) ([]*modules.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*modules.Order
	for _, o := range m.orders {
		if filter.Status != "" && o.Status != filter.Status {
			continue
		}
		cp := *o
		out = append(out, &cp)
	}
	return out, nil
}

func (m *mockOrderRepo) UpdateStatus(_ context.Context, id string, status modules.OrderStatus) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	o, ok := m.orders[id]
	if !ok {
		return usecase.ErrOrderNotFound
	}
	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}

func (m *mockOrderRepo) RecordHistory(_ context.Context, h *modules.OrderHistory) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.history = append(m.history, h)
	return nil
}

func (m *mockOrderRepo) GetHistory(_ context.Context, orderID string) ([]*modules.OrderHistory, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var res []*modules.OrderHistory
	for _, h := range m.history {
		if h.OrderID == orderID {
			res = append(res, h)
		}
	}
	return res, nil
}

func (m *mockOrderRepo) GetMyOrders(_ context.Context, userID string) ([]*modules.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*modules.Order
	for _, o := range m.orders {
		if o.ClientID == userID {
			cp := *o
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *mockOrderRepo) GetByIDs(_ context.Context, ids []string) ([]*modules.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*modules.Order
	for _, id := range ids {
		if o, ok := m.orders[id]; ok {
			cp := *o
			out = append(out, &cp)
		}
	}
	return out, nil
}

func (m *mockOrderRepo) GetByCourierID(_ context.Context, courierID string) ([]*modules.Order, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var out []*modules.Order
	for _, o := range m.orders {
		if o.CourierID == courierID {
			cp := *o
			out = append(out, &cp)
		}
	}
	return out, nil
}

// ── Assignment ────────────────────────────────────────────────────────────────

type mockAssignmentRepo struct {
	assignments map[string]*modules.Assignment
}

func newMockAssignmentRepo() *mockAssignmentRepo {
	return &mockAssignmentRepo{assignments: make(map[string]*modules.Assignment)}
}

func (m *mockAssignmentRepo) Create(_ context.Context, a *modules.Assignment) error {
	m.assignments[a.OrderID] = a
	return nil
}

func (m *mockAssignmentRepo) GetByOrderID(_ context.Context, orderID string) (*modules.Assignment, error) {
	a, ok := m.assignments[orderID]
	if !ok {
		return nil, errors.New("assignment not found")
	}
	return a, nil
}

func (m *mockAssignmentRepo) Complete(_ context.Context, orderID string, completedAt time.Time) error {
	a, ok := m.assignments[orderID]
	if !ok {
		return errors.New("assignment not found")
	}
	a.CompletedAt = &completedAt
	return nil
}

func (m *mockAssignmentRepo) GetByCourierID(_ context.Context, courierID string) ([]*modules.Assignment, error) {
	var out []*modules.Assignment
	for _, a := range m.assignments {
		if a.CourierID == courierID {
			out = append(out, a)
		}
	}
	return out, nil
}
