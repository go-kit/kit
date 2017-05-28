package sterrors_test

import (
	"errors"
	"os"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/sterrors"
)

func ExampleStErrorsWith() {
	a := func() error {
		return errors.New("example error")
	}

	b := func() error {
		err := a()
		if err != nil {
			return sterrors.With(err, "key1", "value1")
		}
		return nil
	}

	c := func() error {
		err := b()
		if err != nil {
			return sterrors.With(err, "key2", "value2")
		}
		return nil
	}

	logger := log.NewLogfmtLogger(os.Stdout)
	err := c().(sterrors.ErrKeyValser)
	if err != nil {
		log.NewContext(logger).With("err", err).Log(err.KeyVals()...)
	}
	// Output: err="example error" key1=value1 key2=value2
}

func ExampleStErrorsWithPrefix() {
	a := func() error {
		return errors.New("example error")
	}

	b := func() error {
		err := a()
		if err != nil {
			return sterrors.With(err, "key1", "value1")
		}
		return nil
	}

	c := func() error {
		err := b()
		if err != nil {
			return sterrors.WithPrefix(err, "key2", "value2")
		}
		return nil
	}

	logger := log.NewLogfmtLogger(os.Stdout)
	err := c().(sterrors.ErrKeyValser)
	if err != nil {
		log.NewContext(logger).With("err", err).Log(err.KeyVals()...)
	}
	// Output: err="example error" key2=value2 key1=value1
}
