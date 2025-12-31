package di

import (
	"sync"

	"go.uber.org/dig"
)

type DigContainer interface {
	ProvideWithName(constructor interface{}, name string) error
	Provide(constructor interface{}, opts ...dig.ProvideOption) error
	MustProvide(constructor interface{}, opts ...dig.ProvideOption)
	Invoke(function interface{}, opts ...dig.InvokeOption) error
	MustInvoke(function interface{}, opts ...dig.InvokeOption)
}

func GetDigContainer() DigContainer {
	initOnce.Do(func() {
		digContainerInstance = &digContainer{
			dig.New(),
		}
	})

	return digContainerInstance
}

var (
	initOnce sync.Once

	digContainerInstance *digContainer = nil
)

type digContainer struct {
	*dig.Container
}

func (c *digContainer) ProvideWithName(constructor interface{}, name string) error {
	return c.Provide(constructor, dig.Name(name))
}

func (c *digContainer) MustProvide(constructor interface{}, opts ...dig.ProvideOption) {
	if err := c.Provide(constructor, opts...); err != nil {
		panic(err)
	}
}

func (c *digContainer) MustInvoke(function interface{}, opts ...dig.InvokeOption) {
	if err := c.Invoke(function, opts...); err != nil {
		panic(err)
	}
}
