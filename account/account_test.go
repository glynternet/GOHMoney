package account_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/glynternet/GOHMoney/account"
	"github.com/glynternet/GOHMoney/common"
	"github.com/glynternet/GOHMoney/money/currency"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	start := time.Now()
	invalidCurrency, err := currency.New("QWERTYUIOP")
	assert.NotNil(t, err)
	a, err := account.New("TEST_ACCOUNT", invalidCurrency, start)
	assert.Equal(t, account.Account{}, a)

	a, err = account.New("TEST_ACCOUNT", newTestCurrency(t, "EUR"), start)
	assert.Nil(t, err)
	assert.Equal(t, currency.Code("EUR"), a.CurrencyCode())
	assert.False(t, a.End().Valid)

	close := start.Add(100 * time.Hour)
	assert.Nil(t, account.CloseTime(close)(&a))
	assert.True(t, a.End().EqualTime(close))
}

func TestAccount_MarshalJSON(t *testing.T) {
	now := time.Now()
	a, err := account.New("TEST ACCOUNT", newTestCurrency(t, "EUR"), now)
	common.FatalIfError(t, err, "Creating Account for testing")
	bytes, err := json.Marshal(&a)
	common.FatalIfError(t, err, "Marshalling json for testing")

	var b account.Account
	err = json.Unmarshal(bytes, &b)
	common.ErrorIfErrorf(t, err, "Unmarshalling Account json")
	assert.True(t, b.Equal(a))

	close := now.Add(48 * time.Hour)
	err = account.CloseTime(close)(&a)
	assert.Nil(t, err)
	assert.True(t, a.End().EqualTime(close))
	bytes, err = json.Marshal(&a)
	common.FatalIfError(t, err, "Marshalling json")

	var c account.Account
	err = json.Unmarshal(bytes, &c)
	common.ErrorIfErrorf(t, err, "Unmarshalling Account json")
	assert.True(t, c.Equal(a), "bytes: %s", bytes)
}

func TestAccount_Equal(t *testing.T) {
	now := time.Now()
	a, err := account.New("A", newTestCurrency(t, "EUR"), now)
	common.ErrorIfError(t, err, "Creating account")
	for _, test := range []struct {
		name       string
		start, end time.Time
		equal      bool
	}{
		{"A", now, time.Time{}, true},
		{"B", now, time.Time{}, false},
		{"A", now.AddDate(-1, 0, 0), time.Time{}, false},
		{"A", now, now.Add(1), false},
		{"A", now.AddDate(-1, 0, 0), now.Add(1), false},
		{"B", now.AddDate(-1, 0, 0), now.Add(1), false},
	} {
		var os []account.Option
		if !test.end.IsZero() {
			os = append(os, account.CloseTime(test.end))
		}
		b, err := account.New(test.name, newTestCurrency(t, "EUR"), test.start, os...)
		if err != nil {
			t.Errorf("Error creating account for testing: %s", err)
		}
		equal := a.Equal(b)
		if equal != test.equal {
			t.Errorf("Expected %t, but got %t.\nA: %v\nB: %v", test.equal, equal, a, b)
		}
	}
}

func newTestCurrency(t *testing.T, code string) currency.Code {
	c, err := currency.New(code)
	common.FatalIfError(t, err, "Creating New Currency Code")
	return c
}
