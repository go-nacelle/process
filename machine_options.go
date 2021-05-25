package process

type MachineConfigFunc func(*machineBuilder)

// WithInjecter configures a machine builder instance to use the given inject hook.
func WithInjecter(injectHook Injecter) MachineConfigFunc {
	return func(b *machineBuilder) { b.injecter = injectHook }
}

// WithHealth configures a machine builder instance to use the given health instance.
func WithHealth(health *Health) MachineConfigFunc {
	return func(b *machineBuilder) { b.health = health }
}
