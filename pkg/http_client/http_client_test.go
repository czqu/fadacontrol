package http_client

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewClientBuilder(t *testing.T) {

	builder := NewClientBuilder()
	assert.NotNil(t, builder)

	assert.Empty(t, builder.proxyURL)

}
