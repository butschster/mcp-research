-- Fresh MySQL 8.4+ schema, equivalent to SQLite migrations 001–027.
-- SQLite retains its upgrade history. Future schema changes need a migration
-- for every supported dialect. Date strings remain UTC for API compatibility.

CREATE TABLE `researches` (
    `id` VARCHAR(255) PRIMARY KEY,
    `name` LONGTEXT NOT NULL,
    `description` LONGTEXT NOT NULL DEFAULT (''),
    `goal` LONGTEXT NOT NULL DEFAULT (''),
    `status` VARCHAR(255) NOT NULL DEFAULT 'active',
    `instruction` LONGTEXT NOT NULL DEFAULT (''),
    `memory` LONGTEXT NOT NULL DEFAULT ('[]'),
    `tags` LONGTEXT NOT NULL DEFAULT ('[]'),
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `code` VARCHAR(255) NOT NULL DEFAULT '',
    `user_id` VARCHAR(255) DEFAULT NULL ,
    `team_id` VARCHAR(255) DEFAULT NULL ,
    `template_slug` LONGTEXT NOT NULL DEFAULT (''),
    `template_version` INTEGER NOT NULL DEFAULT 0,
    code_unique VARCHAR(255) GENERATED ALWAYS AS (NULLIF(code, '')) STORED
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `sections` (
    `id` VARCHAR(255) PRIMARY KEY,
    `research_id` VARCHAR(255) NOT NULL ,
    `name` VARCHAR(255) NOT NULL,
    `display_name` LONGTEXT NOT NULL DEFAULT (''),
    `description` LONGTEXT NOT NULL DEFAULT (''),
    `status` LONGTEXT NOT NULL DEFAULT ('draft'),
    `position` INTEGER NOT NULL DEFAULT 0,
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `code` VARCHAR(255) NOT NULL DEFAULT '',
    `field_spec` LONGTEXT NOT NULL DEFAULT ('[]'),
    `spec_version` INTEGER NOT NULL DEFAULT 0,
    UNIQUE(research_id, name),
    code_unique VARCHAR(255) GENERATED ALWAYS AS (NULLIF(code, '')) STORED
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `entries` (
    `id` VARCHAR(255) PRIMARY KEY,
    `research_id` VARCHAR(255) NOT NULL ,
    `section_id` VARCHAR(255) NOT NULL ,
    `title` LONGTEXT NOT NULL DEFAULT (''),
    `content` LONGTEXT NOT NULL DEFAULT (''),
    `description` LONGTEXT NOT NULL DEFAULT (''),
    `status` LONGTEXT NOT NULL DEFAULT ('draft'),
    `tags` LONGTEXT NOT NULL DEFAULT ('[]'),
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `code` VARCHAR(255) NOT NULL DEFAULT '',
    `session_id` VARCHAR(255) DEFAULT NULL ,
    `entry_type` VARCHAR(255) NOT NULL DEFAULT 'markdown',
    `metadata` LONGTEXT NOT NULL DEFAULT ('{}'),
    `spec_version` INTEGER NOT NULL DEFAULT 0,
    code_unique VARCHAR(255) GENERATED ALWAYS AS (NULLIF(code, '')) STORED
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `sessions` (
    `id` VARCHAR(255) PRIMARY KEY,
    `research_id` VARCHAR(255) NOT NULL ,
    `title` LONGTEXT NOT NULL DEFAULT (''),
    `focus` LONGTEXT NOT NULL DEFAULT (''),
    `status` VARCHAR(255) NOT NULL DEFAULT 'active',
    `notes` LONGTEXT NOT NULL DEFAULT (''),
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `code` VARCHAR(255) NOT NULL DEFAULT '',
    code_unique VARCHAR(255) GENERATED ALWAYS AS (NULLIF(code, '')) STORED
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `questions` (
    `id` VARCHAR(255) PRIMARY KEY,
    `session_id` VARCHAR(255) NOT NULL ,
    `text` LONGTEXT NOT NULL,
    `area` LONGTEXT NOT NULL DEFAULT (''),
    `rationale` LONGTEXT NOT NULL DEFAULT (''),
    `priority` LONGTEXT NOT NULL DEFAULT ('medium'),
    `status` VARCHAR(255) NOT NULL DEFAULT 'pending',
    `answer` LONGTEXT NOT NULL DEFAULT (''),
    `parent_id` VARCHAR(255) DEFAULT NULL,
    `position` INTEGER NOT NULL DEFAULT 0,
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `code` VARCHAR(255) NOT NULL DEFAULT '',
    code_unique VARCHAR(255) GENERATED ALWAYS AS (NULLIF(code, '')) STORED
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `tasks` (
    `id` VARCHAR(255) PRIMARY KEY,
    `research_id` VARCHAR(255) NOT NULL ,
    `title` LONGTEXT NOT NULL,
    `description` LONGTEXT NOT NULL DEFAULT (''),
    `status` VARCHAR(255) NOT NULL DEFAULT 'pending',
    `priority` LONGTEXT NOT NULL DEFAULT ('medium'),
    `result` LONGTEXT NOT NULL DEFAULT (''),
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `completed_at` VARCHAR(32) DEFAULT NULL,
    `code` VARCHAR(255) NOT NULL DEFAULT '',
    code_unique VARCHAR(255) GENERATED ALWAYS AS (NULLIF(code, '')) STORED
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `crossrefs` (
    `id` BIGINT PRIMARY KEY AUTO_INCREMENT,
    `source_type` VARCHAR(255) NOT NULL DEFAULT 'entry',
    `source_id` VARCHAR(255) NOT NULL DEFAULT '',
    `source_entry_id` VARCHAR(255) DEFAULT NULL,
    `source_research_id` VARCHAR(255) NOT NULL,
    `target_entry_id` VARCHAR(255) DEFAULT NULL,
    `target_research_id` VARCHAR(255) DEFAULT NULL,
    `target_ref` LONGTEXT NOT NULL,
    `resolved` INTEGER NOT NULL DEFAULT 0,
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `target_roadmap_id` VARCHAR(255) DEFAULT NULL,
    `target_node_id` VARCHAR(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `users` (
    `id` VARCHAR(255) PRIMARY KEY,
    `email` VARCHAR(255) NOT NULL UNIQUE,
    `password_hash` LONGTEXT NOT NULL,
    `name` LONGTEXT NOT NULL DEFAULT (''),
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `api_keys` (
    `id` VARCHAR(255) PRIMARY KEY,
    `user_id` VARCHAR(255) NOT NULL ,
    `name` LONGTEXT NOT NULL DEFAULT (''),
    `key_hash` VARCHAR(255) NOT NULL UNIQUE,
    `key_prefix` LONGTEXT NOT NULL,
    `last_used_at` VARCHAR(32),
    `expires_at` VARCHAR(32),
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `oauth_clients` (
    `id` VARCHAR(255) PRIMARY KEY,
    `user_id` VARCHAR(255) DEFAULT NULL ,
    `secret_hash` LONGTEXT NOT NULL,
    `name` LONGTEXT NOT NULL,
    `redirect_uris` LONGTEXT NOT NULL DEFAULT ('[]'),
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `oauth_codes` (
    `code` VARCHAR(255) PRIMARY KEY,
    `client_id` VARCHAR(255) NOT NULL ,
    `user_id` VARCHAR(255) NOT NULL ,
    `redirect_uri` LONGTEXT NOT NULL,
    `scope` LONGTEXT NOT NULL DEFAULT (''),
    `expires_at` VARCHAR(32) NOT NULL,
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `code_challenge` LONGTEXT NOT NULL DEFAULT (''),
    `code_challenge_method` LONGTEXT NOT NULL DEFAULT ('')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `oauth_tokens` (
    `id` VARCHAR(255) PRIMARY KEY,
    `client_id` VARCHAR(255) NOT NULL ,
    `user_id` VARCHAR(255) NOT NULL ,
    `access_token_hash` VARCHAR(255) NOT NULL UNIQUE,
    `refresh_token_hash` LONGTEXT NOT NULL DEFAULT (''),
    `scope` LONGTEXT NOT NULL DEFAULT (''),
    `expires_at` VARCHAR(32) NOT NULL,
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `external_links` (
    `id` VARCHAR(255) PRIMARY KEY,
    `source_type` VARCHAR(255) NOT NULL DEFAULT 'entry',
    `source_id` VARCHAR(255) NOT NULL,
    `research_id` VARCHAR(255) NOT NULL ,
    `url` LONGTEXT NOT NULL,
    `title` LONGTEXT NOT NULL DEFAULT (''),
    `domain` VARCHAR(255) NOT NULL DEFAULT '',
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `roadmaps` (
    `id` VARCHAR(255) PRIMARY KEY,
    `code` LONGTEXT NOT NULL DEFAULT (''),
    `research_id` VARCHAR(255) NOT NULL ,
    `title` LONGTEXT NOT NULL,
    `description` LONGTEXT NOT NULL DEFAULT (''),
    `statuses` LONGTEXT NOT NULL DEFAULT ('[]'),
    `status` LONGTEXT NOT NULL DEFAULT ('active'),
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `stages` LONGTEXT NOT NULL DEFAULT ('[]'),
    `view` LONGTEXT NOT NULL DEFAULT ('graph')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `roadmap_nodes` (
    `id` VARCHAR(255) PRIMARY KEY,
    `code` LONGTEXT NOT NULL DEFAULT (''),
    `roadmap_id` VARCHAR(255) NOT NULL ,
    `title` LONGTEXT NOT NULL,
    `description` LONGTEXT NOT NULL DEFAULT (''),
    `node_type` LONGTEXT NOT NULL DEFAULT ('step'),
    `status` LONGTEXT NOT NULL DEFAULT (''),
    `position_x` DOUBLE NOT NULL DEFAULT 0,
    `position_y` DOUBLE NOT NULL DEFAULT 0,
    `parent_id` VARCHAR(255) DEFAULT NULL,
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `ref_type` LONGTEXT,
    `ref_id` VARCHAR(255),
    `metadata` LONGTEXT,
    `stage` LONGTEXT NOT NULL DEFAULT (''),
    `node_date` LONGTEXT NOT NULL DEFAULT (''),
    `node_end_date` LONGTEXT NOT NULL DEFAULT ('')
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `roadmap_edges` (
    `id` VARCHAR(255) PRIMARY KEY,
    `roadmap_id` VARCHAR(255) NOT NULL ,
    `source_node_id` VARCHAR(255) NOT NULL ,
    `target_node_id` VARCHAR(255) NOT NULL ,
    `label` LONGTEXT NOT NULL DEFAULT (''),
    `edge_type` LONGTEXT NOT NULL DEFAULT ('default'),
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `entry_blocks` (
    `entry_id` VARCHAR(255)    NOT NULL ,
    `research_id` VARCHAR(255)    NOT NULL ,
    `block_id` VARCHAR(255)    NOT NULL,
    `position` INTEGER NOT NULL,
    `type` VARCHAR(255)    NOT NULL,
    `data` LONGTEXT    NOT NULL DEFAULT ('{}'),
    `state` LONGTEXT    NOT NULL DEFAULT (''),
    PRIMARY KEY (entry_id, block_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `entry_revisions` (
    `id` VARCHAR(255) PRIMARY KEY,
    `entry_id` VARCHAR(255)    NOT NULL ,
    `research_id` VARCHAR(255)    NOT NULL ,
    `revision` INTEGER NOT NULL,
    `title` LONGTEXT    NOT NULL DEFAULT (''),
    `description` LONGTEXT    NOT NULL DEFAULT (''),
    `content` LONGTEXT    NOT NULL DEFAULT (''),
    `entry_type` LONGTEXT    NOT NULL DEFAULT ('markdown'),
    `status` LONGTEXT    NOT NULL DEFAULT (''),
    `tags` LONGTEXT    NOT NULL DEFAULT ('[]'),
    `author_kind` LONGTEXT    NOT NULL DEFAULT ('agent'),
    `session_id` VARCHAR(255),
    `user_id` VARCHAR(255),
    `summary` LONGTEXT    NOT NULL DEFAULT (''),
    `created_at` VARCHAR(255)    NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `metadata` LONGTEXT NOT NULL DEFAULT ('{}'),
    `spec_version` INTEGER NOT NULL DEFAULT 0,
    UNIQUE(entry_id, revision)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `teams` (
    `id` VARCHAR(255) PRIMARY KEY,
    `name` LONGTEXT NOT NULL,
    `personal` INTEGER NOT NULL DEFAULT 0,
    `created_by` VARCHAR(255) DEFAULT NULL ,
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `team_members` (
    `team_id` VARCHAR(255) NOT NULL ,
    `user_id` VARCHAR(255) NOT NULL ,
    `role` LONGTEXT NOT NULL DEFAULT ('viewer'),
    `invited_by` VARCHAR(255) DEFAULT NULL ,
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    PRIMARY KEY (team_id, user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `team_invites` (
    `id` VARCHAR(255) PRIMARY KEY,
    `team_id` VARCHAR(255) NOT NULL ,
    `email` LONGTEXT NOT NULL DEFAULT (''),
    `role` LONGTEXT NOT NULL DEFAULT ('viewer'),
    `token_hash` VARCHAR(255) NOT NULL UNIQUE,
    `invited_by` VARCHAR(255) NOT NULL ,
    `expires_at` VARCHAR(32) NOT NULL,
    `accepted_at` VARCHAR(32),
    `accepted_by` VARCHAR(255) DEFAULT NULL ,
    `revoked_at` VARCHAR(32),
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `shares` (
    `id` VARCHAR(255) PRIMARY KEY,
    `token_hash` VARCHAR(255) NOT NULL UNIQUE,
    `research_id` VARCHAR(255) NOT NULL ,
    `user_id` VARCHAR(255) DEFAULT NULL ,
    `scope` LONGTEXT NOT NULL DEFAULT ('research'),
    `target_id` VARCHAR(255) NOT NULL DEFAULT '',
    `label` LONGTEXT NOT NULL DEFAULT (''),
    `include` LONGTEXT NOT NULL DEFAULT ('{}'),
    `password_hash` LONGTEXT NOT NULL DEFAULT (''),
    `expires_at` VARCHAR(32),
    `revoked_at` VARCHAR(32),
    `last_seen_at` VARCHAR(32),
    `view_count` INTEGER NOT NULL DEFAULT 0,
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s'))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `skills` (
    `id` VARCHAR(255) PRIMARY KEY,
    `team_id` VARCHAR(255) ,
    `research_id` VARCHAR(255) ,
    `user_id` VARCHAR(255) ,
    `slug` VARCHAR(255) NOT NULL,
    `name` LONGTEXT NOT NULL,
    `description` LONGTEXT NOT NULL DEFAULT (''),
    `body` LONGTEXT NOT NULL DEFAULT (''),
    `tier` LONGTEXT NOT NULL DEFAULT ('team'),
    `ambient` INTEGER NOT NULL DEFAULT 0,
    `forked_from` LONGTEXT,
    `needs_trigger` INTEGER NOT NULL DEFAULT 0,
    `version` INTEGER NOT NULL DEFAULT 1,
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    CHECK (team_id IS NULL OR research_id IS NULL),
    builtin_slug VARCHAR(255) GENERATED ALWAYS AS (CASE WHEN team_id IS NULL AND research_id IS NULL THEN slug ELSE NULL END) VIRTUAL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `research_skills` (
    `research_id` VARCHAR(255) NOT NULL ,
    `skill_id` VARCHAR(255) NOT NULL ,
    `via_template` INTEGER NOT NULL DEFAULT 0,
    `attached_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    PRIMARY KEY (research_id, skill_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `research_templates` (
    `id` VARCHAR(255) PRIMARY KEY,
    `team_id` VARCHAR(255) ,
    `user_id` VARCHAR(255) ,
    `slug` VARCHAR(255) NOT NULL,
    `name` LONGTEXT NOT NULL,
    `description` LONGTEXT NOT NULL DEFAULT (''),
    `when_to_use` LONGTEXT NOT NULL DEFAULT (''),
    `when_not_to_use` LONGTEXT NOT NULL DEFAULT (''),
    `body` LONGTEXT NOT NULL DEFAULT (''),
    `skills` LONGTEXT NOT NULL DEFAULT ('[]'),
    `source` LONGTEXT NOT NULL DEFAULT ('user'),
    `forked_from` LONGTEXT,
    `version` INTEGER NOT NULL DEFAULT 1,
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    global_slug VARCHAR(255) GENERATED ALWAYS AS (CASE WHEN team_id IS NULL THEN slug ELSE NULL END) VIRTUAL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `annotations` (
    `id` VARCHAR(255) PRIMARY KEY,
    `code` LONGTEXT NOT NULL,
    `research_id` VARCHAR(255) NOT NULL ,
    `entry_id` VARCHAR(255) NOT NULL ,
    `block_id` VARCHAR(255) NOT NULL DEFAULT '',
    `quote_exact` LONGTEXT NOT NULL,
    `quote_prefix` LONGTEXT NOT NULL DEFAULT (''),
    `quote_suffix` LONGTEXT NOT NULL DEFAULT (''),
    `anchored_revision` INTEGER NOT NULL DEFAULT 0,
    `kind` LONGTEXT NOT NULL,
    `body` LONGTEXT NOT NULL DEFAULT (''),
    `author_kind` LONGTEXT NOT NULL DEFAULT ('human'),
    `user_id` VARCHAR(255),
    `status` VARCHAR(255) NOT NULL DEFAULT 'open',
    `resolution` LONGTEXT NOT NULL DEFAULT (''),
    `resolved_revision` INTEGER,
    `session_id` VARCHAR(255),
    `task_id` VARCHAR(255),
    `rejections` LONGTEXT    NOT NULL DEFAULT ('[]'),
    `attempts` INTEGER NOT NULL DEFAULT 0,
    `created_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `updated_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    `answered_at` VARCHAR(32),
    `closed_at` VARCHAR(32)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

CREATE TABLE `entry_views` (
    `viewer_id` VARCHAR(255) NOT NULL,
    `user_id` VARCHAR(255) ,
    `entry_id` VARCHAR(255) NOT NULL ,
    `seen_revision` INTEGER NOT NULL CHECK (seen_revision > 0),
    `seen_at` VARCHAR(32) NOT NULL DEFAULT (DATE_FORMAT(UTC_TIMESTAMP(), '%Y-%m-%d %H:%i:%s')),
    PRIMARY KEY (viewer_id, entry_id),
    CHECK (user_id IS NULL OR viewer_id = user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin;

ALTER TABLE `researches` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

ALTER TABLE `researches` ADD FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE;

ALTER TABLE `sections` ADD FOREIGN KEY (`research_id`) REFERENCES `researches` (`id`) ON DELETE CASCADE;

ALTER TABLE `entries` ADD FOREIGN KEY (`research_id`) REFERENCES `researches` (`id`) ON DELETE CASCADE;

ALTER TABLE `entries` ADD FOREIGN KEY (`section_id`) REFERENCES `sections` (`id`) ON DELETE CASCADE;

ALTER TABLE `entries` ADD FOREIGN KEY (`session_id`) REFERENCES `sessions` (`id`) ON DELETE SET NULL;

ALTER TABLE `sessions` ADD FOREIGN KEY (`research_id`) REFERENCES `researches` (`id`) ON DELETE CASCADE;

ALTER TABLE `questions` ADD FOREIGN KEY (`session_id`) REFERENCES `sessions` (`id`) ON DELETE CASCADE;

ALTER TABLE `questions` ADD FOREIGN KEY (parent_id) REFERENCES questions(id) ON DELETE SET NULL;

ALTER TABLE `tasks` ADD FOREIGN KEY (`research_id`) REFERENCES `researches` (`id`) ON DELETE CASCADE;

ALTER TABLE `api_keys` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

ALTER TABLE `oauth_clients` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

ALTER TABLE `oauth_codes` ADD FOREIGN KEY (`client_id`) REFERENCES `oauth_clients` (`id`) ON DELETE CASCADE;

ALTER TABLE `oauth_codes` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

ALTER TABLE `oauth_tokens` ADD FOREIGN KEY (`client_id`) REFERENCES `oauth_clients` (`id`) ON DELETE CASCADE;

ALTER TABLE `oauth_tokens` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

ALTER TABLE `external_links` ADD FOREIGN KEY (`research_id`) REFERENCES `researches` (`id`) ON DELETE CASCADE;

ALTER TABLE `roadmaps` ADD FOREIGN KEY (`research_id`) REFERENCES `researches` (`id`) ON DELETE CASCADE;

ALTER TABLE `roadmap_nodes` ADD FOREIGN KEY (`roadmap_id`) REFERENCES `roadmaps` (`id`) ON DELETE CASCADE;

ALTER TABLE `roadmap_edges` ADD FOREIGN KEY (`roadmap_id`) REFERENCES `roadmaps` (`id`) ON DELETE CASCADE;

ALTER TABLE `roadmap_edges` ADD FOREIGN KEY (`source_node_id`) REFERENCES `roadmap_nodes` (`id`) ON DELETE CASCADE;

ALTER TABLE `roadmap_edges` ADD FOREIGN KEY (`target_node_id`) REFERENCES `roadmap_nodes` (`id`) ON DELETE CASCADE;

ALTER TABLE `entry_blocks` ADD FOREIGN KEY (`entry_id`) REFERENCES `entries` (`id`) ON DELETE CASCADE;

ALTER TABLE `entry_blocks` ADD FOREIGN KEY (`research_id`) REFERENCES `researches` (`id`) ON DELETE CASCADE;

ALTER TABLE `entry_revisions` ADD FOREIGN KEY (`entry_id`) REFERENCES `entries` (`id`) ON DELETE CASCADE;

ALTER TABLE `entry_revisions` ADD FOREIGN KEY (`research_id`) REFERENCES `researches` (`id`) ON DELETE CASCADE;

ALTER TABLE `teams` ADD FOREIGN KEY (`created_by`) REFERENCES `users` (`id`) ON DELETE SET NULL;

ALTER TABLE `team_members` ADD FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE;

ALTER TABLE `team_members` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

ALTER TABLE `team_members` ADD FOREIGN KEY (`invited_by`) REFERENCES `users` (`id`) ON DELETE SET NULL;

ALTER TABLE `team_invites` ADD FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE;

ALTER TABLE `team_invites` ADD FOREIGN KEY (`invited_by`) REFERENCES `users` (`id`) ON DELETE CASCADE;

ALTER TABLE `team_invites` ADD FOREIGN KEY (`accepted_by`) REFERENCES `users` (`id`) ON DELETE SET NULL;

ALTER TABLE `shares` ADD FOREIGN KEY (`research_id`) REFERENCES `researches` (`id`) ON DELETE CASCADE;

ALTER TABLE `shares` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

ALTER TABLE `skills` ADD FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE;

ALTER TABLE `skills` ADD FOREIGN KEY (`research_id`) REFERENCES `researches` (`id`) ON DELETE CASCADE;

ALTER TABLE `skills` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL;

ALTER TABLE `research_skills` ADD FOREIGN KEY (`research_id`) REFERENCES `researches` (`id`) ON DELETE CASCADE;

ALTER TABLE `research_skills` ADD FOREIGN KEY (`skill_id`) REFERENCES `skills` (`id`) ON DELETE CASCADE;

ALTER TABLE `research_templates` ADD FOREIGN KEY (`team_id`) REFERENCES `teams` (`id`) ON DELETE CASCADE;

ALTER TABLE `research_templates` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE SET NULL;

ALTER TABLE `annotations` ADD FOREIGN KEY (`research_id`) REFERENCES `researches` (`id`) ON DELETE CASCADE;

ALTER TABLE `annotations` ADD FOREIGN KEY (`entry_id`) REFERENCES `entries` (`id`) ON DELETE CASCADE;

ALTER TABLE `entry_views` ADD FOREIGN KEY (`user_id`) REFERENCES `users` (`id`) ON DELETE CASCADE;

ALTER TABLE `entry_views` ADD FOREIGN KEY (`entry_id`) REFERENCES `entries` (`id`) ON DELETE CASCADE;

CREATE INDEX idx_researches_status ON researches(status);

CREATE INDEX idx_sections_research ON sections(research_id);

CREATE INDEX idx_entries_section ON entries(section_id);

CREATE INDEX idx_entries_research ON entries(research_id);

CREATE INDEX idx_sessions_research ON sessions(research_id);

CREATE INDEX idx_sessions_status ON sessions(status);

CREATE INDEX idx_questions_session ON questions(session_id);

CREATE INDEX idx_questions_parent ON questions(parent_id);

CREATE INDEX idx_questions_status ON questions(status);

CREATE INDEX idx_tasks_research ON tasks(research_id);

CREATE INDEX idx_tasks_status ON tasks(status);

CREATE UNIQUE INDEX idx_researches_code ON researches(code_unique);

CREATE UNIQUE INDEX idx_entries_code_research ON entries(research_id, code_unique);

CREATE UNIQUE INDEX idx_sections_code_research ON sections(research_id, code_unique);

CREATE UNIQUE INDEX idx_sessions_code_research ON sessions(research_id, code_unique);

CREATE UNIQUE INDEX idx_questions_code_session ON questions(session_id, code_unique);

CREATE UNIQUE INDEX idx_tasks_code_research ON tasks(research_id, code_unique);

CREATE INDEX idx_crossrefs_source ON crossrefs(source_type, source_id);

CREATE INDEX idx_crossrefs_source_research ON crossrefs(source_research_id);

CREATE INDEX idx_crossrefs_target_entry ON crossrefs(target_entry_id);

CREATE INDEX idx_api_keys_user ON api_keys(user_id);

CREATE INDEX idx_researches_user ON researches(user_id);

CREATE INDEX idx_entries_session ON entries(session_id);

CREATE INDEX idx_external_links_research ON external_links(research_id);

CREATE INDEX idx_external_links_source ON external_links(source_type, source_id);

CREATE INDEX idx_external_links_domain ON external_links(domain);

CREATE INDEX idx_roadmaps_research ON roadmaps(research_id);

CREATE INDEX idx_roadmap_nodes_roadmap ON roadmap_nodes(roadmap_id);

CREATE INDEX idx_roadmap_nodes_parent ON roadmap_nodes(parent_id);

CREATE INDEX idx_roadmap_edges_roadmap ON roadmap_edges(roadmap_id);

CREATE INDEX idx_roadmap_edges_source ON roadmap_edges(source_node_id);

CREATE INDEX idx_roadmap_edges_target ON roadmap_edges(target_node_id);

CREATE INDEX idx_crossrefs_target_roadmap ON crossrefs(target_roadmap_id);

CREATE INDEX idx_entries_type ON entries(entry_type);

CREATE INDEX idx_entry_blocks_order ON entry_blocks(entry_id, position);

CREATE INDEX idx_entry_blocks_research ON entry_blocks(research_id, type);

CREATE INDEX idx_entry_revisions_entry ON entry_revisions(entry_id, revision DESC);

CREATE INDEX idx_entry_revisions_session ON entry_revisions(session_id);

CREATE INDEX idx_entry_revisions_research ON entry_revisions(research_id, created_at);

CREATE INDEX idx_team_members_user ON team_members(user_id);

CREATE INDEX idx_team_invites_team ON team_invites(team_id);

CREATE INDEX idx_researches_team ON researches(team_id);

CREATE INDEX idx_shares_research ON shares(research_id);

CREATE UNIQUE INDEX idx_skills_builtin_slug
    ON skills(builtin_slug);

CREATE UNIQUE INDEX idx_skills_team_slug
    ON skills(team_id, slug);

CREATE UNIQUE INDEX idx_skills_private_slug
    ON skills(research_id, slug);

CREATE INDEX idx_research_skills_skill ON research_skills(skill_id);

CREATE UNIQUE INDEX idx_templates_global_slug
    ON research_templates(global_slug);

CREATE UNIQUE INDEX idx_templates_team_slug
    ON research_templates(team_id, slug);

CREATE INDEX idx_annotations_queue ON annotations(research_id, status);

CREATE INDEX idx_annotations_entry ON annotations(entry_id, block_id);

CREATE INDEX idx_entry_views_entry ON entry_views(entry_id);

INSERT INTO teams (id, name, personal) VALUES ('team-local', 'Local', 0);
