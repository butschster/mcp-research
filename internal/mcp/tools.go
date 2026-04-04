package mcp

import (
	"github.com/butschster/mcp-research/internal/mcp/tools"
)

func (s *Server) registerTools() {
	// Research tools
	tools.RegisterResearchCreate(s.server, s.research, s.log)
	tools.RegisterResearchGet(s.server, s.research, s.section, s.session, s.log)
	tools.RegisterResearchList(s.server, s.research, s.log)
	tools.RegisterResearchUpdate(s.server, s.research, s.log)
	tools.RegisterResearchAddSection(s.server, s.research, s.log)

	// Section tools
	tools.RegisterSectionList(s.server, s.section, s.log)
	tools.RegisterSectionUpdate(s.server, s.section, s.log)

	// Entry tools
	tools.RegisterEntryCreate(s.server, s.entry, s.log)
	tools.RegisterEntryList(s.server, s.entry, s.log)
	tools.RegisterEntryRead(s.server, s.entry, s.log)
	tools.RegisterEntryUpdate(s.server, s.entry, s.log)

	// Session tools
	tools.RegisterSessionCreate(s.server, s.session, s.log)
	tools.RegisterSessionGet(s.server, s.session, s.log)
	tools.RegisterSessionUpdate(s.server, s.session, s.log)

	// Question tools
	tools.RegisterQuestionCreate(s.server, s.session, s.log)
	tools.RegisterQuestionUpdate(s.server, s.session, s.log)
	tools.RegisterQuestionList(s.server, s.session, s.log)

	// Task tools
	tools.RegisterTaskCreate(s.server, s.task, s.log)
	tools.RegisterTaskUpdate(s.server, s.task, s.log)
	tools.RegisterTaskList(s.server, s.task, s.log)
	tools.RegisterTaskDelete(s.server, s.task, s.log)
}
