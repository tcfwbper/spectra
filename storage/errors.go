package storage

import "errors"

// ErrNotInitialized indicates that the .spectra directory was not found.
var ErrNotInitialized = errors.New("spectra not initialized: .spectra directory not found")

// ErrSessionDirExists indicates that the session directory already exists.
var ErrSessionDirExists = errors.New("session directory already exists")
