package process

import (
	"github.com/go-nacelle/config"
)

// ConfigurationRegistry is a wrapper around a bare configuration loader object.
type ConfigurationRegistry interface {
	// Register populates the given configuration object. References to the
	// object and tag modifiers may be held for later reporting.
	Register(target interface{}, modifiers ...config.TagModifier)
}

type Configurable interface {
	// RegisterConfiguration is a hook provided with a configuration registry. The
	// registry object populates configuration objects and aggregates configuration
	// and validation errors into a central location.
	//
	// This hook is called prior to the Init method of any registered initializer
	// or process.
	RegisterConfiguration(registry ConfigurationRegistry)
}

func newConfigurationRegistry(c config.Config, meta namedInitializer, f func(registeredConfig)) ConfigurationRegistry {
	return configurationRegisterFunc(func(target interface{}, modifiers ...config.TagModifier) {
		f(registeredConfig{
			meta:      meta,
			target:    target,
			loadErr:   c.Load(target, modifiers...),
			modifiers: modifiers,
		})
	})
}

type configurationRegisterFunc func(target interface{}, modifiers ...config.TagModifier)

func (f configurationRegisterFunc) Register(target interface{}, modifiers ...config.TagModifier) {
	f(target, modifiers...)
}
