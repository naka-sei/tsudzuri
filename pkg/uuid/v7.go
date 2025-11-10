package uuid

import guuid "github.com/google/uuid"

// NewV7 returns a UUIDv7 value or panics if generation fails.
func NewV7() guuid.UUID {
	u, err := guuid.NewV7()
	if err != nil {
		panic(err)
	}
	return u
}
