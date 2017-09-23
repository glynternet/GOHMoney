package GOHMoney

import (
	"bytes"
	"fmt"
	"testing"
	"time"
)

func Test_ValidateAccount(t *testing.T) {
	testSets := []struct {
		insertedAccount Account
		AccountFieldError
	}{
		{
			insertedAccount:   Account{},
			AccountFieldError: AccountFieldError{EmptyNameError, ZeroDateOpenedError},
		},
		{
			insertedAccount: Account{
				Name: "TEST_ACCOUNT",
			},
			AccountFieldError: AccountFieldError{ZeroDateOpenedError},
		},
		{
			insertedAccount: Account{
				timeRange: TimeRange{
					Start: NullTime{
						Valid: true,
						Time:  time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
					},
				},
			},
			AccountFieldError: AccountFieldError{EmptyNameError},
		},
		{
			insertedAccount: Account{
				Name: "TEST_ACCOUNT",
				timeRange: TimeRange{
					Start: NullTime{},
					End: NullTime{
						Valid: true,
					},
				},
			},
			AccountFieldError: AccountFieldError{ZeroDateOpenedError, ZeroValidDateClosedError},
		},
		{
			insertedAccount: Account{
				Name: "TEST_ACCOUNT",
				timeRange: TimeRange{
					Start: NullTime{
						Valid: true,
						Time:  time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
					},
					End: NullTime{
						Valid: true,
						Time:  time.Date(1999, 1, 1, 1, 1, 1, 1, time.UTC),
					},
				},
			},
			AccountFieldError: AccountFieldError{string(DateClosedBeforeDateOpenedError)},
		},
		{
			insertedAccount:   newTestAccount(),
			AccountFieldError: nil,
		},
		{
			insertedAccount: Account{
				Name: "TEST_ACCOUNT",
				timeRange: TimeRange{
					Start: NullTime{
						Valid: true,
						Time:  time.Now(),
					},
					End: NullTime{},
				},
			},
			AccountFieldError: nil,
		},
	}
	for _, testSet := range testSets {
		actual := testSet.insertedAccount.Validate()
		expected := testSet.AccountFieldError
		if !stringSlicesMatch(expected, actual) {
			t.Errorf("Unexpected error.\nExpected: %s\nActual  : %s\nInserted Account: %s", expected, actual, testSet.insertedAccount)
		}
	}
}

func newTestAccount() Account {
	return Account{
		Name: "TEST_ACCOUNT",
		timeRange: TimeRange{
			Start: NullTime{
				Valid: true,
				Time:  time.Date(2000, 1, 1, 1, 1, 1, 1, time.UTC),
			},
			End: NullTime{
				Valid: true,
				Time:  time.Date(2001, 1, 1, 1, 1, 1, 1, time.UTC),
			},
		},
	}
}

func stringSlicesMatch(array1, array2 []string) bool {
	if len(array1) != len(array2) {
		return false
	}
	for i := 0; i < len(array1); i++ {
		if array1[i] != array2[i] {
			return false
		}
	}
	return true
}

func Test_IsOpen(t *testing.T) {
	testSets := []struct {
		Account
		IsOpen bool
	}{
		{
			Account: Account{},
			IsOpen:  true,
		},
		{
			Account: Account{
				timeRange: TimeRange{
					End: NullTime{Valid: false},
				},
			},
			IsOpen: true,
		},
		{
			Account: Account{
				timeRange: TimeRange{
					End: NullTime{Valid: true},
				},
			},
			IsOpen: false,
		},
	}
	for _, testSet := range testSets {
		actual := testSet.Account.IsOpen()
		if actual != testSet.IsOpen {
			t.Errorf("Account IsOpen expected %t, got %t. Account: %v", testSet.IsOpen, actual, testSet.Account)
		}
	}
}

func Test_InvalidAccountValidateBalance(t *testing.T) {
	present := time.Now()
	past := present.AddDate(-1, 0, 0)
	future := present.AddDate(1, 0, 0)

	invalidAccount := Account{
		timeRange: TimeRange{
			Start: NullTime{
				Valid: true,
				Time:  future,
			},
			End: NullTime{
				Valid: true,
				Time:  past,
			},
		},
	}

	accountErr := invalidAccount.Validate()
	balanceErr := invalidAccount.ValidateBalance(Balance{})
	switch {
	case accountErr == nil && balanceErr == nil:
		break
	case accountErr == nil || balanceErr == nil,
		accountErr.Error() != balanceErr.Error():
		t.Errorf("ValidateBalance did not return Account error when given invalid account.\nExpected: %s\nActual  : %s", accountErr, balanceErr)
	}
}

func Test_AccountValidateBalance(t *testing.T) {
	present := time.Now()
	past := present.AddDate(-1, 0, 0)
	future := present.AddDate(1, 0, 0)

	openAccount := Account{
		Name: "Test Account",
		timeRange: TimeRange{
			Start: NullTime{
				Valid: true,
				Time:  present,
			},
			End: NullTime{Valid: false},
		},
	}
	closedAccount := Account{
		Name: "Test Account",
		timeRange: TimeRange{
			Start: NullTime{
				Valid: true,
				Time:  present,
			},
			End: NullTime{
				Valid: true,
				Time:  future,
			},
		},
	}

	pastBalance := Balance{Date: past}
	presentBalance := Balance{Date: present}
	futureBalance := Balance{Date: future}
	testSets := []struct {
		Account
		Balance
		error
	}{
		{
			Account: openAccount,
			Balance: pastBalance,
			error: BalanceDateOutOfAccountTimeRange{
				BalanceDate:      pastBalance.Date,
				AccountTimeRange: openAccount.timeRange,
			},
		},
		{
			Account: openAccount,
			Balance: presentBalance,
			error:   nil,
		},
		{
			Account: openAccount,
			Balance: futureBalance,
			error:   nil,
		},
		{
			Account: closedAccount,
			Balance: pastBalance,
			error: BalanceDateOutOfAccountTimeRange{
				BalanceDate:      pastBalance.Date,
				AccountTimeRange: closedAccount.timeRange,
			},
		},
		{
			Account: closedAccount,
			Balance: presentBalance,
			error:   nil,
		},
		{
			Account: closedAccount,
			Balance: futureBalance,
			error:   nil,
		},
		{
			Account: closedAccount,
			Balance: Balance{
				Date: futureBalance.Date.AddDate(1, 0, 0),
			},
			error: BalanceDateOutOfAccountTimeRange{
				BalanceDate:      futureBalance.Date.AddDate(1, 0, 0),
				AccountTimeRange: closedAccount.timeRange,
			},
		},
	}
	for _, testSet := range testSets {
		err := testSet.Account.ValidateBalance(testSet.Balance)
		if testSet.error == err {
			continue
		}
		testSetTyped, testSetErrorIsType := testSet.error.(BalanceDateOutOfAccountTimeRange)
		actualErrorTyped, actualErrorIsType := err.(BalanceDateOutOfAccountTimeRange)
		if testSetErrorIsType != actualErrorIsType {
			t.Errorf("Expected and resultant errors are differently typed.\nExpected: %s\nActual  : %s", testSet.error, err)
			t.Logf("Account: %s\nBalance: %s", testSet.Account, testSet.Balance)
			continue
		}
		var message bytes.Buffer
		fmt.Fprintf(&message, "Unexpected error.\nExpected: %+v\nActual  : %+v", testSetTyped, actualErrorTyped)
		fmt.Fprintf(&message, "\nExpected error: BalanceDate: %s, AccountTimeRange: %+v", testSetTyped.BalanceDate, testSetTyped.AccountTimeRange)
		fmt.Fprintf(&message, "\nActual error  : BalanceDate: %s, AccountTimeRange: %+v", actualErrorTyped.BalanceDate, actualErrorTyped.AccountTimeRange)
		t.Errorf(message.String())
	}
}

func Test_NewAccount(t *testing.T) {
	now := time.Now()
	type testSet struct {
		name  string
		start time.Time
		end   NullTime
		error
	}
	testSets := []testSet{
		{
			name:  "TEST_ACCOUNT",
			start: now,
			end:   NullTime{},
			error: nil,
		},
		{
			name:  "TEST_ACCOUNT_WITH_ACCOUNT_ERROR",
			start: now,
			end:   NullTime{Valid: true, Time: now.AddDate(0, 0, -1)},
			error: AccountFieldError{DateClosedBeforeDateOpenedError.Error()},
		},
	}
	logTestSet := func(ts testSet) { t.Logf("Start: %s,\tEnd: %s,", ts.start, ts.end) }
	for _, set := range testSets {
		account, err := NewAccount(set.name, set.start, set.end)
		actualFieldError, actualIsTyped := err.(AccountFieldError)
		expectedFieldError, expectedIsTyped := set.error.(AccountFieldError)

		if actualIsTyped != expectedIsTyped {
			t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", set.error, err)
		} else if !actualIsTyped && err != set.error {
			t.Errorf("Unexpected error.\n\tExpected: %s\n\tActual  : %s", set.error, err)
			logTestSet(set)
		} else if actualIsTyped && !actualFieldError.equal(expectedFieldError) {
			t.Errorf("Error is correct type but unexpected contents.\n\tExpected: %s\n\tActual  : %s", expectedFieldError, actualFieldError)
			logTestSet(set)
		}
		if account.Name != set.name {
			t.Errorf("Unexpected name.\n\tExpected: %s\n\tActual  : %s", set.name, account.Name)
			logTestSet(set)
		}
		if !account.timeRange.Start.Valid {
			t.Errorf("Returned invalid start time.")
			logTestSet(set)
		}
		if !account.Start().Equal(set.start) {
			t.Errorf("Unexpected start.\n\tExpected: %s\n\tActual  : %s", set.start, account.Start())
			logTestSet(set)
		}
		if account.End().Valid != set.end.Valid {
			t.Errorf("Unexpected end time validity.\n\tExpected: %s\n\tActual  : %s", account.End().Valid, set.end.Valid)
			logTestSet(set)
		}
		if !account.End().Time.Equal(set.end.Time) {
			t.Errorf("Unexpected end time.\n\tExpected: %s\n\tActual  : %s", set.end.Time, account.End().Time)
			logTestSet(set)
		}
	}
}
