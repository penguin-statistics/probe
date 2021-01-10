package densemver

import (
	"fmt"
	"strconv"
	"testing"
)

func shouldError(t *testing.T, err error) {
	if err == nil {
		t.Error("should have error but got", err)
	} else {
		t.Log("ok: error expected and presented:", err)
	}
}

func TestFromString(t *testing.T) {
	var err error
	t.Run("should error on big semver segments", func(t *testing.T) {
		_, err = FromString("65536.255.255")
		shouldError(t, err)

		_, err = FromString("255.255.256")
		shouldError(t, err)

		_, err = FromString("0.0.256")
		shouldError(t, err)

		_, err = FromString("0.257.1")
		shouldError(t, err)

		_, err = FromString("-1.257.1")
		shouldError(t, err)
	})
	t.Run("should error on malformed semver", func(t *testing.T) {
		_, err = FromString("0.-1.2")
		shouldError(t, err)

		_, err = FromString("-1.-1.2")
		shouldError(t, err)

		_, err = FromString("0...")
		shouldError(t, err)

		_, err = FromString("0.-1.214.")
		shouldError(t, err)
	})

	t.Run("should conform idempotency", func(t *testing.T) {
		testCases := map[string]uint32{
			"0.0.0":         0<<16 + 0<<8 + 0,
			"1.2.3":         1<<16 + 2<<8 + 3,
			"12.34.56":      12<<16 + 34<<8 + 56,
			"84.156.214":    84<<16 + 156<<8 + 214,
			"255.255.255":   255<<16 + 255<<8 + 255,
			"65535.255.255": 65535<<16 + 255<<8 + 255,
		}
		for s, i := range testCases {
			fmt.Println("test case str", s, "bin", strconv.FormatUint(uint64(i), 2))

			ds, err := FromString(s)
			if err != nil {
				t.Fatal("failed to parse string", err)
			}
			if ds.Int() != i {
				t.Fatal("parsed string", s, "expect", i, "got", ds.Int())
			}
			t.Log("ok: parsed", s, "as", ds.Int())

			di, err := FromInt(i)
			if err != nil {
				t.Fatal("failed to parse int", err)
			}
			if di.String() != s {
				t.Fatal("parsed int", i, "expect", s, "got", di.String())
			}
			t.Log("ok: parsed", i, "as", di.String())
		}
	})
}
