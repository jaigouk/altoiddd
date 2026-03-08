package mcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/alty-cli/alty/internal/composition"
	"github.com/alty-cli/alty/internal/discovery/domain"
)

// --- Input structs ---

// GuideStartInput is the typed input for guide_start.
type GuideStartInput struct {
	ReadmeContent string `json:"readme_content" jsonschema:"the README content to start discovery with"`
}

// GuideDetectPersonaInput is the typed input for guide_detect_persona.
type GuideDetectPersonaInput struct {
	SessionID string `json:"session_id" jsonschema:"the discovery session ID"`
	Choice    string `json:"choice" jsonschema:"persona choice: 1=Developer, 2=Product Owner, 3=Domain Expert, 4=Mixed"`
}

// GuideAnswerInput is the typed input for guide_answer.
type GuideAnswerInput struct {
	SessionID  string `json:"session_id" jsonschema:"the discovery session ID"`
	QuestionID string `json:"question_id" jsonschema:"the question ID to answer (e.g. Q1)"`
	Answer     string `json:"answer" jsonschema:"the answer text"`
}

// GuideSkipQuestionInput is the typed input for guide_skip_question.
type GuideSkipQuestionInput struct {
	SessionID  string `json:"session_id" jsonschema:"the discovery session ID"`
	QuestionID string `json:"question_id" jsonschema:"the question ID to skip (e.g. Q1)"`
	Reason     string `json:"reason" jsonschema:"reason for skipping"`
}

// GuideConfirmPlaybackInput is the typed input for guide_confirm_playback.
type GuideConfirmPlaybackInput struct {
	SessionID string `json:"session_id" jsonschema:"the discovery session ID"`
	Confirmed bool   `json:"confirmed" jsonschema:"true to confirm, false to reject"`
}

// GuideCompleteInput is the typed input for guide_complete.
type GuideCompleteInput struct {
	SessionID string `json:"session_id" jsonschema:"the discovery session ID"`
}

// GuideStatusInput is the typed input for guide_status.
type GuideStatusInput struct {
	SessionID string `json:"session_id" jsonschema:"the discovery session ID"`
}

// --- Response helpers ---

func textResult(text string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: text}},
	}, nil, nil
}

func toolError(msg string) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: msg}},
	}, nil, nil
}

func requireSessionID(id string) (*mcp.CallToolResult, any, error) {
	if id == "" {
		return toolError("session_id is required")
	}
	return nil, nil, nil
}

// formatNextQuestion builds the "what to do next" message based on session state.
func formatNextQuestion(session *domain.DiscoverySession) string {
	status := session.Status()

	// 1. PLAYBACK_PENDING → tell client to confirm playback
	if status == domain.StatusPlaybackPending {
		return fmt.Sprintf("Session %s: PLAYBACK_PENDING.\n"+
			"Please confirm the playback summary before continuing.\n"+
			"Use guide_confirm_playback(session_id, confirmed=True) to proceed.",
			session.SessionID())
	}

	// 2. COMPLETED → session done
	if status == domain.StatusCompleted {
		return fmt.Sprintf("Session %s: COMPLETED with %d answers.",
			session.SessionID(), len(session.Answers()))
	}

	// 3. Find next unanswered question from the 10-question catalog
	answeredIDs := make(map[string]bool)
	for _, a := range session.Answers() {
		answeredIDs[a.QuestionID()] = true
	}

	register, hasRegister := session.Register()
	for _, q := range domain.QuestionCatalog() {
		if !answeredIDs[q.ID()] {
			text := q.NonTechnicalText()
			if hasRegister && register == domain.RegisterTechnical {
				text = q.TechnicalText()
			}
			return fmt.Sprintf("Session %s: next question %s (%s phase).\n%s",
				session.SessionID(), q.ID(), q.Phase(), text)
		}
	}

	// 4. All questions answered → tell client to complete
	return fmt.Sprintf("Session %s: all questions answered. Use guide_complete to finish.",
		session.SessionID())
}

// --- Tool handlers ---

func guideStartHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, GuideStartInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input GuideStartInput) (*mcp.CallToolResult, any, error) {
		if strings.TrimSpace(input.ReadmeContent) == "" {
			return toolError("readme_content is required")
		}

		session, err := app.DiscoveryHandler.StartSession(input.ReadmeContent)
		if err != nil {
			return toolError(err.Error())
		}

		return textResult(fmt.Sprintf("Discovery session started.\nsession_id: %s\n"+
			"Next step: detect persona with guide_detect_persona(session_id, choice)\n"+
			"Choices: 1=Developer, 2=Product Owner, 3=Domain Expert, 4=Mixed",
			session.SessionID()))
	}
}

func guideDetectPersonaHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, GuideDetectPersonaInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input GuideDetectPersonaInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}

		session, err := app.DiscoveryHandler.DetectPersona(input.SessionID, input.Choice)
		if err != nil {
			return toolError(err.Error())
		}

		persona, _ := session.Persona()
		register, _ := session.Register()

		return textResult(fmt.Sprintf("Persona detected: %s, register: %s.\n%s",
			persona, register, formatNextQuestion(session)))
	}
}

func guideAnswerHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, GuideAnswerInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input GuideAnswerInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}

		session, err := app.DiscoveryHandler.AnswerQuestion(input.SessionID, input.QuestionID, input.Answer)
		if err != nil {
			return toolError(err.Error())
		}

		return textResult(fmt.Sprintf("Recorded answer for %s.\n%s",
			input.QuestionID, formatNextQuestion(session)))
	}
}

func guideSkipQuestionHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, GuideSkipQuestionInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input GuideSkipQuestionInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}

		session, err := app.DiscoveryHandler.SkipQuestion(input.SessionID, input.QuestionID, input.Reason)
		if err != nil {
			return toolError(err.Error())
		}

		return textResult(fmt.Sprintf("Skipped %s (reason: %s).\n%s",
			input.QuestionID, input.Reason, formatNextQuestion(session)))
	}
}

func guideConfirmPlaybackHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, GuideConfirmPlaybackInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input GuideConfirmPlaybackInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}

		session, err := app.DiscoveryHandler.ConfirmPlayback(input.SessionID, input.Confirmed)
		if err != nil {
			return toolError(err.Error())
		}

		status := "confirmed"
		if !input.Confirmed {
			status = "rejected"
		}

		return textResult(fmt.Sprintf("Playback %s.\n%s",
			status, formatNextQuestion(session)))
	}
}

func guideCompleteHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, GuideCompleteInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input GuideCompleteInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}

		session, err := app.DiscoveryHandler.Complete(input.SessionID)
		if err != nil {
			return toolError(err.Error())
		}

		return textResult(fmt.Sprintf("Discovery session completed.\nsession_id: %s\n"+
			"Answers: %d\nEvents: %d\n"+
			"Next step: use generate_artifacts to produce DDD documents.",
			session.SessionID(), len(session.Answers()), len(session.Events())))
	}
}

func guideStatusHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, GuideStatusInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input GuideStatusInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}

		session, err := app.DiscoveryHandler.GetSession(input.SessionID)
		if err != nil {
			return toolError(err.Error())
		}

		persona, hasPersona := session.Persona()
		personaStr := "not set"
		if hasPersona {
			personaStr = string(persona)
		}

		register, hasRegister := session.Register()
		registerStr := "not set"
		if hasRegister {
			registerStr = string(register)
		}

		answers := session.Answers()
		answeredIDs := make([]string, len(answers))
		for i, a := range answers {
			answeredIDs[i] = a.QuestionID()
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Session: %s\n", session.SessionID())
		fmt.Fprintf(&sb, "Status: %s\n", session.Status())
		fmt.Fprintf(&sb, "Persona: %s\n", personaStr)
		fmt.Fprintf(&sb, "Register: %s\n", registerStr)
		fmt.Fprintf(&sb, "Phase: %s\n", session.CurrentPhase())
		fmt.Fprintf(&sb, "Answered: %d", len(answers))
		if len(answeredIDs) > 0 {
			fmt.Fprintf(&sb, " (%s)", strings.Join(answeredIDs, ", "))
		}
		fmt.Fprintln(&sb)
		fmt.Fprintf(&sb, "Playbacks: %d\n", len(session.PlaybackConfirmations()))
		fmt.Fprintf(&sb, "\n%s", formatNextQuestion(session))

		return textResult(sb.String())
	}
}

// --- Registration ---

// RegisterDiscoveryTools registers all 7 guided discovery MCP tools.
func RegisterDiscoveryTools(server *mcp.Server, app *composition.App) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "guide_start",
		Description: "Start a new DDD discovery session from README content",
	}, guideStartHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "guide_detect_persona",
		Description: "Detect user persona for discovery session (1=Developer, 2=Product Owner, 3=Domain Expert, 4=Mixed)",
	}, guideDetectPersonaHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "guide_answer",
		Description: "Submit an answer to a discovery question",
	}, guideAnswerHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "guide_skip_question",
		Description: "Skip a discovery question with a reason",
	}, guideSkipQuestionHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "guide_confirm_playback",
		Description: "Confirm or reject a playback summary",
	}, guideConfirmPlaybackHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "guide_complete",
		Description: "Complete the discovery session",
	}, guideCompleteHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "guide_status",
		Description: "Get current status of a discovery session",
	}, guideStatusHandler(app))
}
