// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

func init() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
}

func TestLogFullBuffer(t *testing.T) {
	size := 4
	m := model.NewLog(client.NewGVR("fred"), makeLogOpts(size), 10*time.Millisecond)
	m.Init(makeFactory())

	v := newTestView()
	m.AddListener(v)

	data := dao.NewLogItems()
	for i := range 2 * size {
		data.Add(dao.NewLogItemFromString("line" + strconv.Itoa(i)))
		m.Append(data.Items()[i])
	}
	m.Notify()

	assert.Equal(t, 1, v.dataCalled)
	assert.Equal(t, 0, v.clearCalled)
	assert.Equal(t, 0, v.errCalled)
}

func TestLogFilter(t *testing.T) {
	uu := map[string]struct {
		q string
		e int
	}{
		"plain": {
			q: "line-1",
			e: 2,
		},
		"regexp": {
			q: `pod-line-[1-3]{1}`,
			e: 4,
		},
		"invert": {
			q: `!pod-line-1`,
			e: 8,
		},
		"fuzzy": {
			q: `-f po-l1`,
			e: 2,
		},
	}

	size := 10
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			m := model.NewLog(client.NewGVR("fred"), makeLogOpts(size), 10*time.Millisecond)
			m.Init(makeFactory())

			v := newTestView()
			m.AddListener(v)

			m.Filter(u.q)
			data := dao.NewLogItems()
			for i := range size {
				data.Add(dao.NewLogItemFromString(fmt.Sprintf("pod-line-%d", i+1)))
				m.Append(data.Items()[i])
			}

			m.Notify()
			assert.Equal(t, 1, v.dataCalled)
			assert.Equal(t, 1, v.clearCalled)
			assert.Equal(t, 0, v.errCalled)
			assert.Len(t, v.data, u.e)

			m.ClearFilter()
			assert.Equal(t, 2, v.dataCalled)
			assert.Equal(t, 2, v.clearCalled)
			assert.Equal(t, 0, v.errCalled)
			assert.Len(t, v.data, size)
		})
	}
}

func TestLogStartStop(t *testing.T) {
	m := model.NewLog(client.NewGVR("fred"), makeLogOpts(4), 10*time.Millisecond)
	m.Init(makeFactory())

	v := newTestView()
	m.AddListener(v)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	m.Start(ctx)
	data := dao.NewLogItems()
	data.Add(dao.NewLogItemFromString("line1"), dao.NewLogItemFromString("line2"))
	for _, d := range data.Items() {
		m.Append(d)
	}
	m.Notify()
	m.Stop()

	assert.Equal(t, 1, v.dataCalled)
	assert.Equal(t, 0, v.clearCalled)
	assert.Equal(t, 1, v.errCalled)
	assert.Len(t, v.data, 2)
}

func TestLogClear(t *testing.T) {
	m := model.NewLog(client.NewGVR("fred"), makeLogOpts(4), 10*time.Millisecond)
	m.Init(makeFactory())
	assert.Equal(t, "fred", m.GetPath())
	assert.Equal(t, "blee", m.GetContainer())

	v := newTestView()
	m.AddListener(v)

	data := dao.NewLogItems()
	data.Add(dao.NewLogItemFromString("line1"), dao.NewLogItemFromString("line2"))
	for _, d := range data.Items() {
		m.Append(d)
	}
	m.Notify()
	m.Clear()

	assert.Equal(t, 1, v.dataCalled)
	assert.Equal(t, 1, v.clearCalled)
	assert.Equal(t, 0, v.errCalled)
	assert.Empty(t, v.data)
}

func TestLogBasic(t *testing.T) {
	m := model.NewLog(client.NewGVR("fred"), makeLogOpts(2), 10*time.Millisecond)
	m.Init(makeFactory())

	v := newTestView()
	m.AddListener(v)

	data := dao.NewLogItems()
	data.Add(dao.NewLogItemFromString("line1"), dao.NewLogItemFromString("line2"))
	m.Set(data)

	assert.Equal(t, 1, v.dataCalled)
	assert.Equal(t, 1, v.clearCalled)
	assert.Equal(t, 0, v.errCalled)
	ll := make([][]byte, data.Len())
	data.Lines(0, false, ll)
	assert.Equal(t, ll, v.data)
}

func TestLogAppend(t *testing.T) {
	m := model.NewLog(client.NewGVR("fred"), makeLogOpts(4), 5*time.Millisecond)
	m.Init(makeFactory())

	v := newTestView()
	m.AddListener(v)
	items := dao.NewLogItems()
	items.Add(dao.NewLogItemFromString("blah blah"))
	m.Set(items)
	ll := make([][]byte, items.Len())
	items.Lines(0, false, ll)
	assert.Equal(t, ll, v.data)

	data := dao.NewLogItems()
	data.Add(
		dao.NewLogItemFromString("line1"),
		dao.NewLogItemFromString("line2"),
	)
	for _, d := range data.Items() {
		m.Append(d)
	}
	assert.Equal(t, 1, v.dataCalled)
	ll = make([][]byte, items.Len())
	items.Lines(0, false, ll)
	assert.Equal(t, ll, v.data)

	m.Notify()
	assert.Equal(t, 2, v.dataCalled)
	assert.Equal(t, 1, v.clearCalled)
	assert.Equal(t, 0, v.errCalled)
}

func TestLogTimedout(t *testing.T) {
	m := model.NewLog(client.NewGVR("fred"), makeLogOpts(4), 10*time.Millisecond)
	m.Init(makeFactory())

	v := newTestView()
	m.AddListener(v)

	m.Filter("line1")
	data := dao.NewLogItems()
	data.Add(
		dao.NewLogItemFromString("line1"),
		dao.NewLogItemFromString("line2"),
		dao.NewLogItemFromString("line3"),
		dao.NewLogItemFromString("line4"),
	)
	for _, d := range data.Items() {
		m.Append(d)
	}
	m.Notify()
	assert.Equal(t, 1, v.dataCalled)
	assert.Equal(t, 1, v.clearCalled)
	assert.Equal(t, 0, v.errCalled)
	const e = "\x1b[38;5;209ml\x1b[0m\x1b[38;5;209mi\x1b[0m\x1b[38;5;209mn\x1b[0m\x1b[38;5;209me\x1b[0m\x1b[38;5;209m1\x1b[0m"
	assert.Equal(t, e, string(v.data[0]))
}

func TestToggleAllContainers(t *testing.T) {
	opts := makeLogOpts(1)
	opts.DefaultContainer = "duh"
	m := model.NewLog(client.NewGVR(""), opts, 10*time.Millisecond)
	m.Init(makeFactory())
	assert.Equal(t, "blee", m.GetContainer())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	m.ToggleAllContainers(ctx)
	assert.Empty(t, m.GetContainer())
	m.ToggleAllContainers(ctx)
	assert.Equal(t, "blee", m.GetContainer())
}

// ----------------------------------------------------------------------------
// Helpers...

func makeLogOpts(count int) *dao.LogOptions {
	return &dao.LogOptions{
		Path:      "fred",
		Container: "blee",
		Lines:     int64(count),
	}
}

// ----------------------------------------------------------------------------

type testView struct {
	data        [][]byte
	dataCalled  int
	clearCalled int
	errCalled   int
}

func newTestView() *testView {
	return &testView{}
}

func (*testView) LogCanceled() {}
func (*testView) LogStop()     {}
func (*testView) LogResume()   {}

func (t *testView) LogChanged(ll [][]byte) {
	t.data = ll
	t.dataCalled++
}

func (t *testView) LogCleared() {
	t.clearCalled++
	t.data = nil
}

func (t *testView) LogFailed(error) {
	t.errCalled++
}

// ----------------------------------------------------------------------------

type testFactory struct{}

var _ dao.Factory = testFactory{}

func (testFactory) Client() client.Connection {
	return nil
}
func (testFactory) Get(*client.GVR, string, bool, labels.Selector) (runtime.Object, error) {
	return nil, nil
}
func (testFactory) List(*client.GVR, string, bool, labels.Selector) ([]runtime.Object, error) {
	return nil, nil
}
func (testFactory) ForResource(string, *client.GVR) (informers.GenericInformer, error) {
	return nil, nil
}
func (testFactory) CanForResource(string, *client.GVR, []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (testFactory) WaitForCacheSync() {}
func (testFactory) Forwarders() watch.Forwarders {
	return nil
}
func (testFactory) DeleteForwarder(string) {}

func makeFactory() dao.Factory {
	return testFactory{}
}
