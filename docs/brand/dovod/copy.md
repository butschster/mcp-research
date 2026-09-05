# Dovod interface language

Dovod helps people work through a goal with an AI assistant, keep the documents and evidence, and decide what to do next. A project can contain research, planning, interviews, or delivery work. Calling every project a “research” makes that scope harder to understand and produces unnatural English such as “researches”.

## Terms

| Interface | Meaning | Existing API term |
| --- | --- | --- |
| Dovod | The product | — |
| Workspace | The signed-in experience containing teams and projects | — |
| Project / Projects | A goal and its documents, questions, tasks, and sessions | `research` / `researches` |
| Document / Documents | A stored Markdown, structured, or interactive document | `entry` / `entries` |
| Section | A group of documents in a project | `section` |
| Session | A Q&A conversation and its linked work | `session` |
| Memory | Context the AI assistant can use in later work | `memory` |
| Skills | Guidance for how the AI assistant works | `skill` |
| Methodology | A starting guide for a project | `template` |
| AI assistant | The user's connected AI client | `agent` |

Use “research” when describing the activity itself or showing an actual API identifier. Keep API routes, tool names, codes, event types, and stored fields unchanged. For example, the onboarding prompt must still name `research/initialize`; writing `project/initialize` would break the instruction. API reference model names stay aligned with the published specification. Do not rewrite user-authored project names or document content.

## Voice and actions

Use short, concrete instructions. Say what the person can do next and what will appear after it. Avoid suggesting that content will appear automatically before the person starts work with an assistant. Name specific clients only where their setup differs or as examples, not as a requirement for all users.

| Before | After |
| --- | --- |
| Research Projects | Projects |
| No research projects yet | Start your first project |
| Type this into Claude… | Ask your connected AI assistant… |
| All entries | All documents |
| What is this research trying to achieve? | What do you want to find out or decide? |
| Notes shared with the agent through research_get | Context your AI assistant can use when working on this project |
| Rendered / Source · All / Open / Off | Document / Source · Marks: All / Open / Off |

Use the same nouns in navigation, search results, counts, permissions, sharing, export, empty states, and accessible labels. Error messages should explain the next useful step without claiming that a failed request proves a project was deleted.

## Starting with a methodology

The catalogue leads with a name and short description. Usage criteria are available on disclosure; the full instructions open from the name. Each row has a “Copy prompt” action. The prompt points to the current Dovod server's absolute `/llms.txt` address, loads the selected methodology with `template_get`, and supplies its `template_slug` when creating a project. An ID selects the exact methodology when team copies share a slug. The detail page uses the same prompt builder.
