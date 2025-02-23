package tsm

import (
	"slices"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

type MockTmux struct {
	sessions            []string
	active_session      string
	last_killed_session string
}

func (tmux *MockTmux) TmuxListSessions() []string {
	return tmux.sessions
}

func (tmux *MockTmux) TmuxKillSession(session string) error {
	var alive_sessions []string
	for _, alive_session := range tmux.sessions {
		if alive_session != session {
			alive_sessions = append(alive_sessions, alive_session)
		} else {
			tmux.last_killed_session = alive_session
		}
	}
	tmux.sessions = alive_sessions

	return nil
}

func (tmux *MockTmux) TmuxSwitchSession(session string) error {
	tmux.active_session = session
	return nil
}

func (tmux *MockTmux) TmuxCreateSession(session string) error {
	tmux.sessions = append(tmux.sessions, session)
	return nil
}

func (tmux *MockTmux) TmuxRenameSession(oldSession string, session string) error {
	idx := slices.Index(tmux.sessions, oldSession)
	tmux.sessions[idx] = session
	return nil
}

func TestCursorMovedInRightDirectionInManageState(t *testing.T) {
	tests := []struct {
		initial_pos  int
		expected_pos int
		n_emit       int
		key_runes    []rune
	}{
		{0, 0, 1, []rune{'c', 't', 'r', 'l', '+', 'p'}},
		{0, 0, 2, []rune{'c', 't', 'r', 'l', '+', 'p'}},
		{1, 0, 1, []rune{'c', 't', 'r', 'l', '+', 'p'}},
		{5, 2, 3, []rune{'c', 't', 'r', 'l', '+', 'p'}},
		{0, 0, 1, []rune{'k'}},
		{0, 0, 2, []rune{'k'}},
		{1, 0, 1, []rune{'k'}},
		{5, 2, 3, []rune{'k'}},
		{0, 1, 1, []rune{'c', 't', 'r', 'l', '+', 'n'}},
		{0, 2, 2, []rune{'c', 't', 'r', 'l', '+', 'n'}},
		{0, 2, 3, []rune{'c', 't', 'r', 'l', '+', 'n'}},
		{0, 2, 4, []rune{'c', 't', 'r', 'l', '+', 'n'}},
		{2, 2, 1, []rune{'c', 't', 'r', 'l', '+', 'n'}},
		{2, 2, 2, []rune{'c', 't', 'r', 'l', '+', 'n'}},
		{0, 1, 1, []rune{'j'}},
		{0, 2, 2, []rune{'j'}},
		{0, 2, 3, []rune{'j'}},
		{0, 2, 4, []rune{'j'}},
		{2, 2, 1, []rune{'j'}},
		{2, 2, 2, []rune{'j'}},
	}

	for _, test := range tests {
		test_model := InitialSessionModel(&MockTmux{sessions: []string{"test_session_1", "test_session_2", "test_session_3"}})
		test_model.cursor = test.initial_pos
		test_model.state = MANAGE_STATE
		for i := 0; i < test.n_emit; i++ {
			msg := tea.Key{Type: tea.KeyRunes, Runes: test.key_runes}
			updModel, _ := test_model.Update(tea.KeyMsg(msg))
			test_model = updModel.(model)
		}
		if test_model.cursor != test.expected_pos {
			t.Errorf("Expected cursor to be %d, got %d", test.expected_pos, test_model.cursor)
		}
	}
}

func TestTmuxKillSessionInManageState(t *testing.T) {
	tests := []struct {
		kill_sessions                []int
		expected_last_killed_session string
		expected_sessions            []string
	}{
		{
			[]int{0},
			"test_session_1",
			[]string{"test_session_2", "test_session_3"},
		},
		{
			[]int{0, 0}, // NOTE: since the list is always moved left
			"test_session_2",
			[]string{"test_session_3"},
		},
		{
			[]int{0, 0, 0}, // NOTE: since the list is always moved left
			"test_session_3",
			[]string{},
		},
		{
			[]int{2},
			"test_session_3",
			[]string{"test_session_1", "test_session_2"},
		},
		{
			[]int{1},
			"test_session_2",
			[]string{"test_session_1", "test_session_3"},
		},
		{
			[]int{2, 1, 0},
			"test_session_1",
			[]string{},
		},
	}

	for _, test := range tests {
		test_model := InitialSessionModel(&MockTmux{sessions: []string{"test_session_1", "test_session_2", "test_session_3"}})
		test_model.state = MANAGE_STATE
		for _, kill_session_cursor := range test.kill_sessions {
			test_model.cursor = kill_session_cursor
			msg := tea.Key{Type: tea.KeyRunes, Runes: []rune{'d'}}
			updModel, _ := test_model.Update(tea.KeyMsg(msg))
			test_model = updModel.(model)
		}
		mockTmux := test_model.tmux.(*MockTmux)
		if mockTmux.last_killed_session != test.expected_last_killed_session {
			t.Errorf("Expected last killed session to be %s, got %s", test.expected_last_killed_session, mockTmux.last_killed_session)
		}
		if len(mockTmux.sessions) != len(test.expected_sessions) {
			t.Errorf("Expected sessions to be %v, got %v", test.expected_sessions, mockTmux.sessions)
		}
		for i := range len(test.expected_sessions) {
			if mockTmux.sessions[i] != test.expected_sessions[i] {
				t.Errorf("Expected session %s, got %s", test.expected_sessions[i], mockTmux.sessions[i])
			}
		}
	}
}

func TestTmuxSwitchSessionInManageState(t *testing.T) {
	tests := []struct {
		switch_sessions         []int
		expected_active_session string
	}{
		{
			[]int{0},
			"test_session_1",
		},
		{
			[]int{1},
			"test_session_2",
		},
		{
			[]int{2},
			"test_session_3",
		},
		{
			[]int{0, 1},
			"test_session_2",
		},
		{
			[]int{0, 1, 2},
			"test_session_3",
		},
	}

	for _, test := range tests {
		test_model := InitialSessionModel(&MockTmux{sessions: []string{"test_session_1", "test_session_2", "test_session_3"}})
		test_model.state = MANAGE_STATE

		for _, switch_session := range test.switch_sessions {
			test_model.cursor = switch_session
			msg := tea.Key{Type: tea.KeyRunes, Runes: []rune{'e', 'n', 't', 'e', 'r'}}
			updModel, _ := test_model.Update(tea.KeyMsg(msg))
			test_model = updModel.(model)
		}
		mockTmux := test_model.tmux.(*MockTmux)
		if mockTmux.active_session != test.expected_active_session {
			t.Errorf("Expected active session to be %s, got %s", test.expected_active_session, mockTmux.active_session)
		}
	}
}

func TestTransitionToCreateState(t *testing.T) {
	test_model := InitialSessionModel(&MockTmux{sessions: []string{"test_session_1", "test_session_2", "test_session_3"}})
	test_model.state = MANAGE_STATE
	msg := tea.Key{Type: tea.KeyRunes, Runes: []rune{'c'}}
	updModel, _ := test_model.Update(tea.KeyMsg(msg))

	castedModel := updModel.(model)
	if castedModel.state != CREATE_STATE {
		t.Errorf("Expected state to be 1, got %d", castedModel.state)
	}
	if castedModel.focused != NEW_SESSION_INPUT {
		t.Errorf("Expected focused to be 0, got %d", castedModel.focused)
	}
}

func TestTransitionToRenameState(t *testing.T) {
	test_model := InitialSessionModel(&MockTmux{sessions: []string{"test_session_1", "test_session_2", "test_session_3"}})
	test_model.state = MANAGE_STATE
	msg := tea.Key{Type: tea.KeyRunes, Runes: []rune{'r'}}
	updModel, _ := test_model.Update(tea.KeyMsg(msg))

	castedModel := updModel.(model)
	if castedModel.state != RENAME_STATE {
		t.Errorf("Expected state to be 2, got %d", castedModel.state)
	}
	if castedModel.focused != RENAME_SESSION_INPUT {
		t.Errorf("Expected focused to be 1, got %d", castedModel.focused)
	}
}

func TestTransitionToFilteringInManageState(t *testing.T) {
	test_model := InitialSessionModel(&MockTmux{sessions: []string{"test_session_1", "test_session_2", "test_session_3"}})
	test_model.state = MANAGE_STATE
	msg := tea.Key{Type: tea.KeyRunes, Runes: []rune{'/'}}
	updModel, _ := test_model.Update(tea.KeyMsg(msg))

	castedModel := updModel.(model)
	if castedModel.state != MANAGE_STATE {
		t.Errorf("Expected state to be 0, got %d", castedModel.state)
	}
	if castedModel.filtering != true {
		t.Errorf("Expected filtering to be true, got %v", castedModel.filtering)
	}
}

func TestQuitApp(t *testing.T) {
	tests := [][]rune{{'q'}, {'c', 't', 'r', 'l', '+', 'c'}}

	for _, test := range tests {
		test_model := InitialSessionModel(&MockTmux{sessions: []string{"test_session_1", "test_session_2", "test_session_3"}})
		test_model.state = MANAGE_STATE
		msg := tea.Key{Type: tea.KeyRunes, Runes: test}
		_, cmd := test_model.Update(tea.KeyMsg(msg))

		if _, ok := interface{}(cmd()).(tea.QuitMsg); !ok {
			t.Fatalf("Expected cmd to be tea.Quit, got %v", cmd())
		}
	}
}

func TestUpdateFilteringInput(t *testing.T) {
	tests := []struct {
		input_updates            [][]rune
		expected_filtering_input string
		expected_filtering       bool
	}{
		{
			[][]rune{{'t'}, {'e'}, {'s'}, {'t'}},
			"test",
			true,
		},
		{
			[][]rune{{'t'}, {'e'}, {'s'}, {'t'}, {'e', 'n', 't', 'e', 'r'}}, // TODO: here need to additinally check "choices"
			"",
			false,
		},
		{
			[][]rune{{'t'}, {'e'}, {'s'}, {'t'}, {'e', 's', 'c'}},
			"",
			false,
		},
	}

	for _, test := range tests {
		test_model := InitialSessionModel(&MockTmux{sessions: []string{"test_session_1", "test_session_2", "test_session_3"}})
		test_model.state = MANAGE_STATE
		test_model.filtering = true

		var updModel tea.Model = test_model
		for _, input := range test.input_updates {
			msg := tea.Key{Type: tea.KeyRunes, Runes: input}
			updModel, _ = updModel.Update(tea.KeyMsg(msg))
		}
		if updModel.(model).filtering_input.Value() != test.expected_filtering_input {
			t.Fatalf("Expected filtering input to be %s, got %s", test.expected_filtering_input, updModel.(model).filtering_input.Value())
		}
		if updModel.(model).filtering != test.expected_filtering {
			t.Fatalf("Expected filtering to be %v, got %v", test.expected_filtering, updModel.(model).filtering)
		}
	}
}
