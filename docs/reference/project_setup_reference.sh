#!/bin/bash
# Original project_setup script - kept as reference for alty init design
# See the actual implementation at bin/alty
#
# Key patterns to preserve:
# - Tool detection and installation (beads, context7)
# - IDE config generation (Claude Code, Cursor, Roo Code)
# - .gitignore management (ensure_gitignore_entry, fix_gitignore_entry)
# - AGENTS.md generation
# - Global config awareness (~/.claude/CLAUDE.md)
# - Summary report at the end
#
# Key changes for alty:
# - Preview-first, confirm before action
# - Never overwrite existing files
# - Conflict rename with _alty suffix
# - Global settings detection and conflict reporting
# - Branch for existing projects (--existing)
# - Test verification for existing projects (zero regression gate)
# - Drop grepai, notebooklm
# - Add .alty/ directory, knowledge base, templates
