package storage

// MaxPayloadSize is the maximum allowed size (in bytes) for any single serialized
// payload persisted by storage components. Both EventStore and SessionMetadataStore
// reference this constant to enforce size limits before writing.
const MaxPayloadSize = 10 * 1024 * 1024
