package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// Loader handles configuration loading and refreshing.
type Loader struct {
	mu           sync.RWMutex
	configFile   string
	cfg          interface{} // Pointer to the config struct
	onUpdateFunc func(interface{})
	stopChan     chan struct{}
}

// Option allows configuring the Loader.
type Option func(*Loader)

// WithFile specifies the configuration file path.
func WithFile(path string) Option {
	return func(l *Loader) {
		l.configFile = path
	}
}

// NewLoader creates a new configuration loader.
func NewLoader(cfg interface{}, opts ...Option) *Loader {
	l := &Loader{
		cfg:      cfg,
		stopChan: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(l)
	}
	return l
}

// Load loads the configuration from defaults, file, and environment variables.
func (l *Loader) Load() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 1. Process Defaults
	if err := processDefaults(l.cfg); err != nil {
		return fmt.Errorf("failed to process defaults: %w", err)
	}

	// 2. Load File (if specified and exists)
	if l.configFile != "" {
		if err := loadFile(l.configFile, l.cfg); err != nil {
			return fmt.Errorf("failed to load config file: %w", err)
		}
	}

	// 3. Process Environment Variables
	if err := processEnv(l.cfg); err != nil {
		return fmt.Errorf("failed to process env vars: %w", err)
	}

	return nil
}

func (l *Loader) MustLoad() error {
	if err := l.Load(); err != nil {
		panic(err)
	}
	return nil
}

// StartAutoRefresh starts a periodic refresh of the configuration.
// It runs in a background goroutine.
func (l *Loader) StartAutoRefresh(interval time.Duration, onUpdate func(interface{})) {
	l.mu.Lock()
	l.onUpdateFunc = onUpdate
	l.mu.Unlock()

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-l.stopChan:
				return
			case <-ticker.C:
				l.refresh()
			}
		}
	}()
}

// StopAutoRefresh stops the background refresh goroutine.
func (l *Loader) StopAutoRefresh() {
	close(l.stopChan)
}

func (l *Loader) refresh() {
	// Create a new instance of the config struct to avoid modifying the current one in place
	// while it might be in use, although for this simple implementation we might just reload
	// into a new struct and then callback.

	// Reflection to create a new pointer to the same type as l.cfg
	t := reflect.TypeOf(l.cfg)
	if t.Kind() != reflect.Ptr {
		return // Should not happen if Load() succeeded
	}

	newCfgVal := reflect.New(t.Elem())
	newCfg := newCfgVal.Interface()

	// 1. Defaults
	if err := processDefaults(newCfg); err != nil {
		// Log error? For now, just ignore failed refresh
		return
	}

	// 2. File
	if l.configFile != "" {
		if err := loadFile(l.configFile, newCfg); err != nil {
			return
		}
	}

	// 3. Env
	if err := processEnv(newCfg); err != nil {
		return
	}

	// Notify
	l.mu.RLock()
	callback := l.onUpdateFunc
	l.mu.RUnlock()

	if callback != nil {
		callback(newCfg)
	}
}

// processDefaults sets default values defined in `default` tag.
func processDefaults(ptr interface{}) error {
	v := reflect.ValueOf(ptr).Elem()
	return setDefaults(v)
}

func setDefaults(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		// Handle recursion for nested structs
		if fieldVal.Kind() == reflect.Struct {
			if err := setDefaults(fieldVal); err != nil {
				return err
			}
			continue
		} else if fieldVal.Kind() == reflect.Ptr && fieldVal.Elem().Kind() == reflect.Struct {
			// Initialize pointer if nil and it has defaults?
			// For simplicity, skip nil pointers or initialize them?
			// Let's skip nil pointers for now unless we want to allocate everything.
			if !fieldVal.IsNil() {
				if err := setDefaults(fieldVal.Elem()); err != nil {
					return err
				}
			}
			continue
		}

		defaultVal := fieldType.Tag.Get("default")
		if defaultVal != "" && isZero(fieldVal) {
			if err := setValue(fieldVal, defaultVal); err != nil {
				return fmt.Errorf("failed to set default for field %s: %w", fieldType.Name, err)
			}
		}
	}
	return nil
}

// loadFile reads and parses the config file.
func loadFile(path string, ptr interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist and we just want to use defaults/env, maybe ignore?
		// Requirement implies "Support file config", usually if file is specified but missing, it's an error.
		return err
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".json":
		return json.Unmarshal(data, ptr)
	case ".yaml", ".yml":
		return yaml.Unmarshal(data, ptr)
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}
}

// processEnv sets values from environment variables defined in `env` tag.
func processEnv(ptr interface{}) error {
	v := reflect.ValueOf(ptr).Elem()
	return setEnv(v)
}

func setEnv(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		fieldVal := v.Field(i)
		fieldType := t.Field(i)

		if !fieldVal.CanSet() {
			continue
		}

		// Handle recursion
		if fieldVal.Kind() == reflect.Struct {
			if err := setEnv(fieldVal); err != nil {
				return err
			}
			continue
		} else if fieldVal.Kind() == reflect.Ptr && !fieldVal.IsNil() && fieldVal.Elem().Kind() == reflect.Struct {
			if err := setEnv(fieldVal.Elem()); err != nil {
				return err
			}
			continue
		}

		envKey := fieldType.Tag.Get("env")
		if envKey != "" {
			val := os.Getenv(envKey)
			if val != "" {
				if err := setValue(fieldVal, val); err != nil {
					return fmt.Errorf("failed to set env %s for field %s: %w", envKey, fieldType.Name, err)
				}
			}
		}
	}
	return nil
}

// setValue converts string to the field's type and sets it.
func setValue(v reflect.Value, s string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			// Try parsing duration if it fails as int
			if v.Type() == reflect.TypeOf(time.Duration(0)) {
				d, err := time.ParseDuration(s)
				if err != nil {
					return err
				}
				v.SetInt(int64(d))
				return nil
			}
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		i, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(i)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)
	case reflect.Slice:
		// Simple comma-separated support for slices?
		// Requirement didn't specify, but it's common.
		// For now, let's assume if it's a slice we split by comma.
		if v.Type().Elem().Kind() == reflect.String {
			parts := strings.Split(s, ",")
			slice := reflect.MakeSlice(v.Type(), len(parts), len(parts))
			for i, part := range parts {
				slice.Index(i).SetString(strings.TrimSpace(part))
			}
			v.Set(slice)
		}
		// Add more slice types if needed
	}
	return nil
}

func isZero(v reflect.Value) bool {
	return v.IsZero()
}
