-- Add short codes to all remaining entities
-- S1, S2 (sections, per research)
-- SS1, SS2 (sessions, per research)
-- Q1, Q2 (questions, per session)
-- T1, T2 (tasks, per research)

ALTER TABLE sections ADD COLUMN code TEXT NOT NULL DEFAULT '';
ALTER TABLE sessions ADD COLUMN code TEXT NOT NULL DEFAULT '';
ALTER TABLE questions ADD COLUMN code TEXT NOT NULL DEFAULT '';
ALTER TABLE tasks ADD COLUMN code TEXT NOT NULL DEFAULT '';

CREATE UNIQUE INDEX IF NOT EXISTS idx_sections_code_research ON sections(research_id, code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_code_research ON sessions(research_id, code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_questions_code_session ON questions(session_id, code) WHERE code != '';
CREATE UNIQUE INDEX IF NOT EXISTS idx_tasks_code_research ON tasks(research_id, code) WHERE code != '';
