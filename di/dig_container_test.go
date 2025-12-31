package di

import (
	"sync"
	"testing"

	"go.uber.org/dig"
)

type TestService struct {
	Name string
}

func TestGetDigContainer(t *testing.T) {
	container1 := GetDigContainer()
	container2 := GetDigContainer()

	if container1 == nil {
		t.Fatal("GetDigContainer should return non-nil container")
	}

	if container1 != container2 {
		t.Fatal("GetDigContainer should return the same instance")
	}
}

func TestDigContainer_Provide(t *testing.T) {
	container := &digContainer{
		Container: dig.New(),
	}

	var receivedService *TestService
	err := container.Provide(func() *TestService {
		return &TestService{Name: "test"}
	})
	if err != nil {
		t.Fatalf("Provide failed: %v", err)
	}

	err = container.Invoke(func(s *TestService) {
		receivedService = s
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if receivedService == nil {
		t.Fatal("Service should not be nil")
	}

	if receivedService.Name != "test" {
		t.Fatalf("Expected name 'test', got '%s'", receivedService.Name)
	}
}

func TestDigContainer_ProvideWithName(t *testing.T) {
	container := &digContainer{
		Container: dig.New(),
	}

	err := container.ProvideWithName(func() *TestService {
		return &TestService{Name: "named_service"}
	}, "myservice")
	if err != nil {
		t.Fatalf("ProvideWithName failed: %v", err)
	}

	type NamedServiceParams struct {
		dig.In
		Service *TestService `name:"myservice"`
	}

	var s *TestService
	err = container.Invoke(func(p NamedServiceParams) {
		s = p.Service
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if s == nil {
		t.Fatal("Named service should not be nil")
	}

	if s.Name != "named_service" {
		t.Fatalf("Expected name 'named_service', got '%s'", s.Name)
	}
}

func TestDigContainer_MustProvide(t *testing.T) {
	container := &digContainer{
		Container: dig.New(),
	}

	var receivedService *TestService
	container.MustProvide(func() *TestService {
		return &TestService{Name: "mustprovide"}
	})

	err := container.Invoke(func(s *TestService) {
		receivedService = s
	})
	if err != nil {
		t.Fatalf("Invoke after MustProvide failed: %v", err)
	}

	if receivedService == nil {
		t.Fatal("Service should not be nil")
	}

	if receivedService.Name != "mustprovide" {
		t.Fatalf("Expected name 'mustprovide', got '%s'", receivedService.Name)
	}
}

func TestDigContainer_MustProvide_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustProvide should panic on error")
		}
	}()

	container := &digContainer{
		Container: dig.New(),
	}

	type Result struct {
		dig.Out
		Service1 *TestService `name:"s1"`
		Service2 *TestService `name:"s2"`
	}

	container.MustProvide(func() Result {
		return Result{
			Service1: &TestService{Name: "test"},
			Service2: &TestService{Name: "test"},
		}
	}, dig.Name("invalid"))
}

func TestDigContainer_MustInvoke(t *testing.T) {
	container := &digContainer{
		Container: dig.New(),
	}

	var invoked bool
	container.MustProvide(func() *TestService {
		return &TestService{Name: "test"}
	})

	container.MustInvoke(func(s *TestService) {
		invoked = true
	})

	if !invoked {
		t.Fatal("Function should have been invoked")
	}
}

func TestDigContainer_MustInvoke_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustInvoke should panic on error")
		}
	}()

	container := &digContainer{
		Container: dig.New(),
	}

	container.MustInvoke(func(s *TestService) {
		_ = s.Name
	})
}

func TestDigContainer_Invoke_Error(t *testing.T) {
	container := &digContainer{
		Container: dig.New(),
	}

	err := container.Invoke(func(s *TestService) {
		_ = s
	})

	if err == nil {
		t.Fatal("Invoke should return error when dependency not provided")
	}
}

func TestDigContainer_ProvideWithOptions(t *testing.T) {
	container := &digContainer{
		Container: dig.New(),
	}

	type NamedResult1 struct {
		dig.Out
		Service1 *TestService `name:"service1"`
	}
	type NamedResult2 struct {
		dig.Out
		Service2 *TestService `name:"service2"`
	}

	err := container.Provide(func() NamedResult1 {
		return NamedResult1{
			Service1: &TestService{Name: "service1"},
		}
	})
	if err != nil {
		t.Fatalf("Provide failed: %v", err)
	}

	err = container.Provide(func() NamedResult2 {
		return NamedResult2{
			Service2: &TestService{Name: "service2"},
		}
	})
	if err != nil {
		t.Fatalf("Provide failed: %v", err)
	}

	type Params struct {
		dig.In
		S1 *TestService `name:"service1"`
		S2 *TestService `name:"service2"`
	}

	var s1, s2 *TestService
	err = container.Invoke(func(p Params) {
		s1 = p.S1
		s2 = p.S2
	})
	if err != nil {
		t.Fatalf("Invoke failed: %v", err)
	}

	if s1 == nil || s2 == nil {
		t.Fatal("Services should not be nil")
	}

	if s1.Name != "service1" || s2.Name != "service2" {
		t.Fatalf("Unexpected service names")
	}
}

func resetDigContainerInstance() {
	initOnce = sync.Once{}
	digContainerInstance = nil
}
