package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	challengedomain "github.com/alto-cli/alto/internal/challenge/domain"
	"github.com/alto-cli/alto/internal/composition"
	"github.com/alto-cli/alto/internal/shared/domain/ddd"
)

// --- Input structs ---

// ChallengeStartInput is the typed input for challenge_start.
type ChallengeStartInput struct {
	MaxPerType *int `json:"max_per_type,omitempty" jsonschema:"maximum challenges per type (default: 2)"`
}

// ChallengeRespondInput is the typed input for challenge_respond.
type ChallengeRespondInput struct {
	SessionID       string   `json:"session_id" jsonschema:"the challenge session ID"`
	ChallengeID     string   `json:"challenge_id" jsonschema:"the challenge ID to respond to (e.g. c0)"`
	Accepted        bool     `json:"accepted" jsonschema:"true to accept the challenge, false to reject"`
	UserResponse    string   `json:"user_response" jsonschema:"explanation of the response"`
	ArtifactUpdates []string `json:"artifact_updates,omitempty" jsonschema:"list of DDD.md updates prompted by this challenge"`
}

// ChallengeStatusInput is the typed input for challenge_status.
type ChallengeStatusInput struct {
	SessionID string `json:"session_id" jsonschema:"the challenge session ID"`
}

// ChallengeCompleteInput is the typed input for challenge_complete.
type ChallengeCompleteInput struct {
	SessionID string  `json:"session_id" jsonschema:"the challenge session ID"`
	DDDPath   *string `json:"ddd_path,omitempty" jsonschema:"path to DDD.md to version (optional)"`
}

// --- Tool handlers ---

func challengeStartHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, ChallengeStartInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input ChallengeStartInput) (*mcp.CallToolResult, any, error) {
		maxPerType := 2 // sensible default
		if input.MaxPerType != nil && *input.MaxPerType > 0 {
			maxPerType = *input.MaxPerType
		}

		// Create a minimal domain model for challenge generation
		// In practice, this would come from a discovery session or parsed DDD.md
		model := ddd.NewDomainModel("challenge-session")

		session, err := app.ChallengeHandler.StartChallenge(ctx, model, maxPerType)
		if err != nil {
			return toolError(err.Error())
		}

		challenges := session.Challenges()
		var sb strings.Builder
		fmt.Fprintf(&sb, "Challenge session started.\n")
		fmt.Fprintf(&sb, "session_id: %s\n", session.SessionID())
		fmt.Fprintf(&sb, "challenges: %d\n", len(challenges))

		if len(challenges) > 0 {
			fmt.Fprintln(&sb, "\nChallenges:")
			ids := session.ChallengeIDs()
			for i, c := range challenges {
				fmt.Fprintf(&sb, "  %s [%s]: %s\n", ids[i], c.ChallengeType(), c.QuestionText())
			}
			fmt.Fprintln(&sb, "\nNext step: use challenge_respond to answer each challenge.")
		} else {
			fmt.Fprintln(&sb, "\nNo challenges generated. Domain model may be complete.")
		}

		return textResult(sb.String())
	}
}

func challengeRespondHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, ChallengeRespondInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input ChallengeRespondInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}
		if strings.TrimSpace(input.ChallengeID) == "" {
			return toolError("challenge_id is required")
		}

		err := app.ChallengeHandler.RespondToChallenge(
			input.SessionID,
			input.ChallengeID,
			input.UserResponse,
			input.Accepted,
			input.ArtifactUpdates,
		)
		if err != nil {
			return toolError(err.Error())
		}

		session, err := app.ChallengeHandler.GetSession(input.SessionID)
		if err != nil {
			return toolError(err.Error())
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "Response recorded for %s.\n", input.ChallengeID)
		fmt.Fprintf(&sb, "Accepted: %v\n", input.Accepted)

		responses := session.Responses()
		challenges := session.Challenges()
		fmt.Fprintf(&sb, "Progress: %d/%d challenges answered\n", len(responses), len(challenges))

		if session.Status() == challengedomain.SessionStatusCompleted {
			fmt.Fprintln(&sb, "\nAll challenges answered. Use challenge_complete to finalize.")
		} else {
			// Show remaining challenges
			answeredIDs := make(map[string]bool)
			for _, r := range responses {
				answeredIDs[r.ChallengeID()] = true
			}
			ids := session.ChallengeIDs()
			for i, c := range challenges {
				if !answeredIDs[ids[i]] {
					fmt.Fprintf(&sb, "\nNext: %s [%s]: %s", ids[i], c.ChallengeType(), c.QuestionText())
					break
				}
			}
		}

		return textResult(sb.String())
	}
}

func challengeStatusHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, ChallengeStatusInput) (*mcp.CallToolResult, any, error) {
	return func(_ context.Context, _ *mcp.CallToolRequest, input ChallengeStatusInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}

		session, err := app.ChallengeHandler.GetSession(input.SessionID)
		if err != nil {
			return toolError(err.Error())
		}

		challenges := session.Challenges()
		responses := session.Responses()

		var sb strings.Builder
		fmt.Fprintf(&sb, "Session: %s\n", session.SessionID())
		fmt.Fprintf(&sb, "Status: %s\n", session.Status())
		fmt.Fprintf(&sb, "Challenges: %d\n", len(challenges))
		fmt.Fprintf(&sb, "Responses: %d\n", len(responses))

		if len(challenges) > 0 {
			fmt.Fprintln(&sb, "\nChallenge details:")
			ids := session.ChallengeIDs()
			answeredIDs := make(map[string]bool)
			for _, r := range responses {
				answeredIDs[r.ChallengeID()] = true
			}
			for i, c := range challenges {
				status := "pending"
				if answeredIDs[ids[i]] {
					status = "answered"
				}
				fmt.Fprintf(&sb, "  %s [%s] (%s): %s\n", ids[i], c.ChallengeType(), status, truncate(c.QuestionText(), 60))
			}
		}

		return textResult(sb.String())
	}
}

func challengeCompleteHandler(app *composition.App) func(context.Context, *mcp.CallToolRequest, ChallengeCompleteInput) (*mcp.CallToolResult, any, error) {
	return func(ctx context.Context, _ *mcp.CallToolRequest, input ChallengeCompleteInput) (*mcp.CallToolResult, any, error) {
		if r, m, e := requireSessionID(input.SessionID); r != nil {
			return r, m, e
		}

		iteration, err := app.ChallengeHandler.CompleteSession(input.SessionID)
		if err != nil {
			return toolError(err.Error())
		}

		responses := iteration.Responses()
		accepted := 0
		rejected := 0
		for _, r := range responses {
			if r.Accepted() {
				accepted++
			} else {
				rejected++
			}
		}

		var sb strings.Builder
		fmt.Fprintln(&sb, "Challenge session completed.")
		fmt.Fprintf(&sb, "session_id: %s\n", input.SessionID)
		fmt.Fprintf(&sb, "challenges: %d\n", len(iteration.Challenges()))
		fmt.Fprintf(&sb, "responses: %d (accepted: %d, rejected: %d)\n", len(responses), accepted, rejected)
		fmt.Fprintf(&sb, "convergence_delta: %d\n", iteration.ConvergenceDelta())

		// Version DDD.md if path provided
		if input.DDDPath != nil && *input.DDDPath != "" {
			if err := app.VersionHandler.VersionDDDDocument(
				ctx,
				*input.DDDPath,
				"challenge",
				iteration.ConvergenceDelta(),
				time.Now(),
			); err != nil {
				fmt.Fprintf(&sb, "\nWarning: failed to version DDD.md: %v", err)
			} else {
				fmt.Fprintf(&sb, "\nDDD.md versioned successfully at %s", *input.DDDPath)
			}
		} else if iteration.ConvergenceDelta() > 0 {
			fmt.Fprintln(&sb, "\nDDD model improvements suggested. Consider updating DDD.md.")
		}

		return textResult(sb.String())
	}
}

// truncate truncates a string to maxLen characters, adding "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// --- Registration ---

// RegisterChallengeTools registers all 4 challenge MCP tools.
func RegisterChallengeTools(server *mcp.Server, app *composition.App) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "challenge_start",
		Description: "Start a new DDD challenge session to probe the domain model for gaps",
	}, challengeStartHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "challenge_respond",
		Description: "Respond to a challenge question (accept or reject with explanation)",
	}, challengeRespondHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "challenge_status",
		Description: "Get current status of a challenge session",
	}, challengeStatusHandler(app))

	mcp.AddTool(server, &mcp.Tool{
		Name:        "challenge_complete",
		Description: "Complete the challenge session and get convergence summary",
	}, challengeCompleteHandler(app))
}
