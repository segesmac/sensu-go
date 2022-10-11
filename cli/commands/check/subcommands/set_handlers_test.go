package subcommands

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSetHandlersCommand(t *testing.T) {
	testCases := []struct {
		testName       string
		args           []string
		fetchResponse  error
		updateResponse error
		expectedOutput string
		expectError    bool
	}{
		{"no args", []string{}, nil, nil, "Usage", true},
		{"fetch error", []string{"checky", "foo"}, fmt.Errorf("error"), nil, "", true},
		{"update error", []string{"checky", "bar"}, nil, fmt.Errorf("error"), "", true},
		{"invalid input", []string{"checky"}, nil, nil, "", true},
		{"valid input", []string{"checky", "handler1,handler2"}, nil, nil, "Updated", false},
	}

	for _, tc := range testCases {
		var name string
		if len(tc.args) > 0 {
			name = tc.args[0]
		}

		t.Run(tc.testName, func(t *testing.T) {
			check := corev2.FixtureCheckConfig("checky")
			cli := test.NewMockCLI()

			client := cli.Client.(*client.MockClient)
			client.On(
				"FetchCheck",
				name,
			).Return(check, tc.fetchResponse)

			client.On(
				"UpdateCheck",
				mock.Anything,
			).Return(tc.updateResponse)

			cmd := SetHandlersCommand(cli)
			out, err := test.RunCmd(cmd, tc.args)
			if tc.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Regexp(t, tc.expectedOutput, out)
		})
	}
}
