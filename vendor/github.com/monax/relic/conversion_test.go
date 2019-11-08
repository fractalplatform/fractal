package relic

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsDate(t *testing.T) {
	_, err := AsDate(time.Now())
	require.NoError(t, err)

	// Support single format
	d, err := AsDate("2018-08-14")
	require.NoError(t, err)
	assert.Equal(t, "14 Aug 18 00:00 +0000", d.Format(time.RFC822Z))

	// Should fails
	_, err = AsDate("20180814")
	require.Error(t, err)
}
