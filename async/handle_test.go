package async

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"testing"

	"github.com/crochee/lirity/routine"
)

type testError struct {
}

func (t testError) SafeCopy() Executor {
	return t
}

func (t testError) ID() string {
	return ""
}

func (t testError) Run(ctx context.Context, data []byte) error {
	fmt.Println("run testError")
	return errors.New("testError failed")
}

type test struct {
}

func (t test) SafeCopy() Executor {
	return t
}

func (t test) ID() string {
	return ""
}

func (t test) Run(ctx context.Context, data []byte) error {
	fmt.Println("90")
	return nil
}

type test1 struct {
	i uint
}

func (t *test1) SafeCopy() Executor {
	tmp := *t
	return &tmp
}

func (t test1) ID() string {
	return ""
}

func (t *test1) Run(ctx context.Context, data []byte) error {
	t.i++
	fmt.Printf("91\t %#v\n", t)
	return nil
}

type multiTest struct {
	list []Executor
}

func (m *multiTest) SafeCopy() Executor {
	tmp := &multiTest{list: make([]Executor, 0, len(m.list))}
	for _, e := range m.list {
		tmp.list = append(tmp.list, e.SafeCopy())
	}
	return tmp
}

func (t multiTest) ID() string {
	return ""
}

func (m *multiTest) Run(ctx context.Context, data []byte) error {
	fmt.Println("mt", len(m.list))
	g := routine.NewGroup(ctx)
	for _, e := range m.list {
		tmp := e
		g.Go(func(ctx context.Context) error {
			return tmp.Run(ctx, nil)
		})
	}
	return g.Wait()
}

func createImage(interface{}) error {
	return nil
}

func TestFunc(t *testing.T) {
	t.Log(runtime.FuncForPC(reflect.ValueOf(createImage).Pointer()).Name())
}

func TestRetry(t *testing.T) {
	m := NewManager()
	if err := m.Register(test{}, &test1{}, &multiTest{list: []Executor{test{}, &test1{}}}); err != nil {
		t.Fatal(err)
	}
	t.Log(m.Run(context.Background(), &Param{
		Name: "async.test",
		Data: nil,
	}))
	t.Log(m.Run(context.Background(), &Param{
		Name: "async.test1",
		Data: nil,
	}))
	t.Log(m.Run(context.Background(), &Param{
		Name: "async.multiTest",
		Data: nil,
	}))
	t.Log(m.Run(context.Background(), &Param{
		Name: "async.multiTest",
		Data: nil,
	}))
}
