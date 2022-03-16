package async

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/crochee/lirity/routine"
)

type testError struct {
}

func (t testError) Run(ctx context.Context, param *Param) error {
	fmt.Println("run testError")
	return errors.New("testError failed")
}

type test struct {
}

func (t test) Run(ctx context.Context, param *Param) error {
	fmt.Println("90")
	return nil
}

type test1 struct {
	i uint
}

func (t *test1) Run(ctx context.Context, param *Param) error {
	t.i++
	fmt.Printf("91\t %#v\n", t)
	return nil
}

type multiTest struct {
	list []Callback
}

func (m *multiTest) Run(ctx context.Context, param *Param) error {
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

func TestRetry(t *testing.T) {
	m := NewManager()
	m.Register("test", test{})
	m.Register("test1", &test1{})
	m.Register("multiTest", &multiTest{list: []Callback{test{}, &test1{}}})

	t.Log(m.Run(context.Background(), &Param{
		TaskType: "test",
		Data:     nil,
	}))
	t.Log(m.Run(context.Background(), &Param{
		TaskType: "test1",
		Data:     nil,
	}))
	t.Log(m.Run(context.Background(), &Param{
		TaskType: "multiTest",
		Data:     nil,
	}))
	t.Log(m.Run(context.Background(), &Param{
		TaskType: "async.multiTest",
		Data:     nil,
	}))
}
