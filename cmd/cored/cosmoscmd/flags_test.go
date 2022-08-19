package cosmoscmd

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/stretchr/testify/assert"
)

func TestModifyArgs(t *testing.T) {
	testCases := []struct {
		input    []string
		flag     string
		newVal   string
		expected []string
		hasErr   bool
	}{
		{
			input:    []string{"cored", "--someFlag=1/2"},
			flag:     flags.FlagHome,
			newVal:   "3",
			expected: []string{"cored", "--someFlag=1/2"},
			hasErr:   true,
		},
		{
			input:    []string{"cored", "--someFlag=1/2"},
			flag:     "someFlag",
			newVal:   "3",
			expected: []string{"cored", "--someFlag=1/2/3"},
		},
		{
			input:    []string{"cored", "--home=1/2", "--chain-id=ch1"},
			flag:     flags.FlagHome,
			newVal:   "3",
			expected: []string{"cored", "--home=1/2/3", "--chain-id=ch1"},
		},
		{
			input:    []string{"cored", "--home=1/2"},
			flag:     flags.FlagHome,
			newVal:   "3",
			expected: []string{"cored", "--home=1/2/3"},
		},
		{
			input:    []string{"cored", "--home", "1/2"},
			flag:     flags.FlagHome,
			newVal:   "3",
			expected: []string{"cored", "--home", "1/2/3"},
		},
		{
			input:    []string{"cored", "--home=1/2/"},
			flag:     flags.FlagHome,
			newVal:   "3",
			expected: []string{"cored", "--home=1/2/3"},
		},
		{
			input:    []string{"cored", "--home", "1/2/"},
			flag:     flags.FlagHome,
			newVal:   "3",
			expected: []string{"cored", "--home", "1/2/3"},
		},
	}

	for tn := range testCases {
		tc := testCases[tn]
		t.Run("", func(t *testing.T) {
			assertT := assert.New(t)
			err := appendStringFlag(tc.input, tc.flag, tc.newVal)
			if tc.hasErr {
				assertT.Error(err)
			} else {
				assertT.NoError(err)
			}
			assertT.EqualValues(tc.expected, tc.input)
		})
	}
}
