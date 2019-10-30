package retry

import (
	"fmt"
	"testing"
)

func TestBackoffBuilder(t *testing.T) {
	fixedBackoff, _ := NewFixedBackoff(123)

	{
		builder := NewBackoffBuilder()
		if _, err := builder.Build(); err == nil {
			t.FailNow()
		}
	}

	{
		builder := NewBackoffBuilder().BaseBackoffSpec("fixed=456")
		if b, err := builder.Build(); err != nil || b == nil {
			t.FailNow()
		} else {
			for i := 0; i < 10000; i++ {
				if b.NextDelayMillis(i) != 456 {
					t.FailNow()
				}
			}
		}
	}

	{
		// error spec
		builder := NewBackoffBuilder().BaseBackoffSpec("fixe=456")
		if b, err := builder.Build(); err == nil || b != nil {
			t.FailNow()
		}
	}

	{
		builder := NewBackoffBuilder().BaseBackoff(NoDelayBackoff)

		if b, err := builder.Build(); err != nil || b == nil {
			t.FailNow()
		}
	}

	{
		builder := NewBackoffBuilder().
			BaseBackoff(NoDelayBackoff).
			WithJitter(0.9)

		if _, err := builder.Build(); err != nil {
			t.FailNow()
		}
	}

	{
		builder := NewBackoffBuilder().
			BaseBackoff(NoDelayBackoff).
			WithJitterBound(0.9, 1.2)

		if _, err := builder.Build(); err == nil {
			t.FailNow()
		}
	}

	{
		builder := NewBackoffBuilder().
			BaseBackoff(NoDelayBackoff).
			WithJitter(0.9).
			WithJitterBound(0.9, 1.2)

		if _, err := builder.Build(); err == nil {
			t.FailNow()
		}
	}

	{
		builder := NewBackoffBuilder().
			BaseBackoff(fixedBackoff).
			WithLimit(5).
			WithJitter(0.9).
			WithJitterBound(0.9, 1.2)

		if _, err := builder.Build(); err == nil {
			t.FailNow()
		}
	}

	{
		builder := NewBackoffBuilder().
			BaseBackoff(fixedBackoff).
			WithLimit(5)

		if b, err := builder.Build(); err != nil {
			t.FailNow()
		} else {
			for i := 0; i < 100; i++ {
				d := b.NextDelayMillis(i)
				if i < 5 && d != 123 {
					fmt.Println(i, d)
					t.FailNow()
				}

				if i >= 5 && d != -1 {
					fmt.Println(i, d)
					t.FailNow()
				}
			}
		}
	}

	{
		builder := NewBackoffBuilder().
			BaseBackoff(fixedBackoff).
			WithLimit(-1).
			WithJitter(0.9).
			WithJitterBound(0.9, 1.2)

		if _, err := builder.Build(); err == nil {
			t.FailNow()
		}
	}
}

func TestParseSpec(t *testing.T) {
	if b, err := parseFromSpec("dummy"); err == nil || b != nil {
		t.FailNow()
	}

	{
		// test exponential
		if _, err := parseFromSpec("exponential="); err != ErrInvalidSpecFormat {
			t.FailNow()
		}

		if _, err := parseFromSpec("exponential=1:"); err != ErrInvalidSpecFormat {
			t.FailNow()
		}

		if _, err := parseFromSpec("exponential=1:2"); err != ErrInvalidSpecFormat {
			t.FailNow()
		}

		if _, err := parseFromSpec("exponential=a:2:3"); err == nil {
			t.FailNow()
		}

		if _, err := parseFromSpec("exponential=1:a:3"); err == nil {
			t.FailNow()
		}

		if _, err := parseFromSpec("exponential=1:2:a"); err == nil {
			t.FailNow()
		}

		if b, err := parseFromSpec("exponential=1:2:3"); err != nil {
			t.Error(err)
			t.FailNow()
		} else if tmp := b.(*ExponentialBackoff); tmp.initialDelayMillis != 1 || tmp.maxDelayMillis != 2 || tmp.multiplier != 3 {
			t.FailNow()
		}

		if b, err := parseFromSpec("exponential=:201:3"); err != nil {
			t.Error(err)
			t.FailNow()
		} else if tmp := b.(*ExponentialBackoff); tmp.initialDelayMillis != DefaultDelayMillis || tmp.maxDelayMillis != 201 || tmp.multiplier != 3 {
			t.FailNow()
		}

		if b, err := parseFromSpec("exponential=::3"); err != nil {
			t.Error(err)
			t.FailNow()
		} else if tmp := b.(*ExponentialBackoff); tmp.initialDelayMillis != DefaultDelayMillis || tmp.maxDelayMillis != DefaultMaxDelayMillis || tmp.multiplier != 3 {
			t.FailNow()
		}

		if b, err := parseFromSpec("exponential=::"); err != nil {
			t.Error(err)
			t.FailNow()
		} else if tmp := b.(*ExponentialBackoff); tmp.initialDelayMillis != DefaultDelayMillis || tmp.maxDelayMillis != DefaultMaxDelayMillis || tmp.multiplier != DefaultMultiplier {
			t.FailNow()
		}
	}

	{
		// test fixed
		if b, err := parseFromSpec("fixed="); err != nil {
			t.FailNow()
		} else if tmp := b.(*FixedBackoff); tmp.delayMillis != DefaultDelayMillis {
			t.FailNow()
		}

		// test fixed
		if _, err := parseFromSpec("fixed=a"); err == nil {
			t.FailNow()
		}

		if b, err := parseFromSpec("fixed=123"); err != nil {
			t.FailNow()
		} else if tmp := b.(*FixedBackoff); tmp.delayMillis != 123 {
			t.FailNow()
		}
	}

	{
		// test random
		if _, err := parseFromSpec("random="); err != ErrInvalidSpecFormat {
			t.FailNow()
		}

		if _, err := parseFromSpec("random=1"); err != ErrInvalidSpecFormat {
			t.FailNow()
		}

		if _, err := parseFromSpec("random=1:2"); err != nil {
			t.FailNow()
		}

		if _, err := parseFromSpec("random=a:2"); err == nil {
			t.FailNow()
		}

		if _, err := parseFromSpec("random=1:a"); err == nil {
			t.FailNow()
		}

		if b, err := parseFromSpec("random=:"); err != nil {
			t.Error(err)
			t.FailNow()
		} else if tmp := b.(*RandomBackoff); tmp.minDelayMillis != DefaultMinDelayMillis || tmp.maxDelayMillis != DefaultMaxDelayMillis {
			t.FailNow()
		}

		if b, err := parseFromSpec("random=12:"); err != nil {
			t.Error(err)
			t.FailNow()
		} else if tmp := b.(*RandomBackoff); tmp.minDelayMillis != 12 || tmp.maxDelayMillis != DefaultMaxDelayMillis {
			t.FailNow()
		}

		if b, err := parseFromSpec("random=:1234"); err != nil {
			t.Error(err)
			t.FailNow()
		} else if tmp := b.(*RandomBackoff); tmp.minDelayMillis != DefaultMinDelayMillis || tmp.maxDelayMillis != 1234 {
			t.FailNow()
		}

		if b, err := parseFromSpec("random=12:1234"); err != nil {
			t.Error(err)
			t.FailNow()
		} else if tmp := b.(*RandomBackoff); tmp.minDelayMillis != 12 || tmp.maxDelayMillis != 1234 {
			t.FailNow()
		}
	}
}