package tui

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/abiswas97/sentei/internal/pipeline"
	"github.com/abiswas97/sentei/internal/repo"
)

type repoEventMsg pipeline.Event

type repoDoneMsg struct {
	result interface{} // CreateResult, CloneResult, or MigrateResult
}

// startRepoPipeline launches the appropriate pipeline based on opts type.
func (m *Model) startRepoPipeline(opts interface{}) tea.Cmd {
	ch := make(chan pipeline.Event, 50)
	resultCh := make(chan interface{}, 1)
	m.repo.eventCh = ch
	m.repo.resultCh = resultCh

	runner := m.runner
	shell := m.shell

	switch o := opts.(type) {
	case repo.CreateOptions:
		go func() {
			result := repo.Create(runner, shell, o, func(e pipeline.Event) { ch <- e })
			close(ch)
			resultCh <- result
		}()
	case repo.CloneOptions:
		go func() {
			result := repo.Clone(runner, o, func(e pipeline.Event) { ch <- e })
			close(ch)
			resultCh <- result
		}()
	case repo.MigrateOptions:
		go func() {
			result := repo.Migrate(runner, shell, o, func(e pipeline.Event) { ch <- e })
			close(ch)
			resultCh <- result
		}()
	}

	return m.waitForRepoEvent()
}

func (m Model) waitForRepoEvent() tea.Cmd {
	return func() tea.Msg {
		ev, ok := <-m.repo.eventCh
		if !ok {
			result := <-m.repo.resultCh
			return repoDoneMsg{result: result}
		}
		return repoEventMsg(ev)
	}
}

func (m Model) updateRepoProgress(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, keys.Quit) {
			return m, tea.Quit
		}
		return m, nil

	case repoEventMsg:
		m.repo.events = append(m.repo.events, pipeline.Event(msg))
		return m, m.waitForRepoEvent()

	case repoDoneMsg:
		m.repo.result = msg.result
		targetView := repoSummaryView
		if m.repo.opType == "migrate" {
			targetView = migrateSummaryView
		}
		return m.holdOrAdvance(targetView)
	}
	return m, nil
}

func (m Model) viewRepoProgress() string {
	var title, subject string
	switch m.repo.opType {
	case "create":
		title = "Creating Repository"
		subject = m.repo.nameInput.Value()
	case "clone":
		title = "Cloning Repository"
		subject = m.repo.urlInput.Value()
	case "migrate":
		title = "Migrating Repository"
		subject = m.repoPath
	}

	return ProgressLayout{
		Title:    title,
		Subtitle: subject,
		Phases:   buildPhaseDisplays(m.repo.events),
		Width:    m.width,
		Height:   m.height,
		Hints:    progressFooter,
	}.View()
}
