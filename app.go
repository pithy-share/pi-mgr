package main

import (
	"context"
)

// App struct holds application-level state
type App struct {
	ctx context.Context
}

// NewApp creates a new App instance
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	// Ensure storage directory exists
	ensureConfigDir()
}