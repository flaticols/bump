package cmd

import (
	"testing"

	"github.com/flaticols/bump/semver"
	"github.com/stretchr/testify/require"
)

func TestBuildTag_WithAndWithoutPrefix(t *testing.T) {
	v, err := semver.Parse("1.2.3")
	require.NoError(t, err)

	optsNoPrefix := &Options{Prefix: "", P: TextPrinters{Version: func(s string) string { return "v" + s }}}
	optsWithPrefix := &Options{Prefix: "pkg/x/", P: TextPrinters{Version: func(s string) string { return "v" + s }}}

	require.Equal(t, "v1.2.3", buildTag(optsNoPrefix, v))
	require.Equal(t, "pkg/x/v1.2.3", buildTag(optsWithPrefix, v))
}
