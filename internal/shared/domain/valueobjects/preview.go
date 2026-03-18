package valueobjects

// FileActionType is the kind of file operation alto plans to perform.
type FileActionType string

// File action type constants.
const (
	FileActionCreate         FileActionType = "create"
	FileActionSkip           FileActionType = "skip"
	FileActionConflictRename FileActionType = "conflict_rename"
)

// AllFileActionTypes returns all valid file action type values.
func AllFileActionTypes() []FileActionType {
	return []FileActionType{FileActionCreate, FileActionSkip, FileActionConflictRename}
}

// FileAction is a single planned file operation.
type FileAction struct {
	path        string
	actionType  FileActionType
	reason      string
	renamedPath string
}

// NewFileAction creates a FileAction value object.
func NewFileAction(path string, actionType FileActionType, reason string, renamedPath string) FileAction {
	return FileAction{
		path:        path,
		actionType:  actionType,
		reason:      reason,
		renamedPath: renamedPath,
	}
}

// Path returns the file path.
func (fa FileAction) Path() string { return fa.path }

// ActionType returns the action type.
func (fa FileAction) ActionType() FileActionType { return fa.actionType }

// Reason returns the human-readable reason.
func (fa FileAction) Reason() string { return fa.reason }

// RenamedPath returns the renamed path (empty if not a conflict rename).
func (fa FileAction) RenamedPath() string { return fa.renamedPath }

// GlobalSettingConflict is a conflict between a global tool setting and local project config.
type GlobalSettingConflict struct {
	tool        string
	globalPath  string
	globalValue string
	localValue  string
}

// NewGlobalSettingConflict creates a GlobalSettingConflict value object.
func NewGlobalSettingConflict(tool, globalPath, globalValue, localValue string) GlobalSettingConflict {
	return GlobalSettingConflict{
		tool:        tool,
		globalPath:  globalPath,
		globalValue: globalValue,
		localValue:  localValue,
	}
}

// Tool returns the tool identifier.
func (c GlobalSettingConflict) Tool() string { return c.tool }

// GlobalPath returns the path to the global settings file.
func (c GlobalSettingConflict) GlobalPath() string { return c.globalPath }

// GlobalValue returns the current global value.
func (c GlobalSettingConflict) GlobalValue() string { return c.globalValue }

// LocalValue returns the desired local value.
func (c GlobalSettingConflict) LocalValue() string { return c.localValue }

// ConflictResolution describes how to resolve a GlobalSettingConflict.
type ConflictResolution string

// Conflict resolution constants.
const (
	ConflictKeepGlobal          ConflictResolution = "keep_global"
	ConflictUpdateGlobal        ConflictResolution = "update_global"
	ConflictSetLocalWithWarning ConflictResolution = "set_local_with_warning"
)

// Preview is an immutable snapshot of all planned bootstrap actions.
type Preview struct {
	fileActions          []FileAction
	conflicts            []GlobalSettingConflict
	conflictDescriptions []string
}

// NewPreview creates a Preview value object.
func NewPreview(fileActions []FileAction, conflicts []GlobalSettingConflict, conflictDescriptions []string) Preview {
	fa := make([]FileAction, len(fileActions))
	copy(fa, fileActions)
	c := make([]GlobalSettingConflict, len(conflicts))
	copy(c, conflicts)
	cd := make([]string, len(conflictDescriptions))
	copy(cd, conflictDescriptions)
	return Preview{
		fileActions:          fa,
		conflicts:            c,
		conflictDescriptions: cd,
	}
}

// FileActions returns a defensive copy of file actions.
func (p Preview) FileActions() []FileAction {
	out := make([]FileAction, len(p.fileActions))
	copy(out, p.fileActions)
	return out
}

// Conflicts returns a defensive copy of conflicts.
func (p Preview) Conflicts() []GlobalSettingConflict {
	out := make([]GlobalSettingConflict, len(p.conflicts))
	copy(out, p.conflicts)
	return out
}

// ConflictDescriptions returns a defensive copy of conflict descriptions.
func (p Preview) ConflictDescriptions() []string {
	out := make([]string, len(p.conflictDescriptions))
	copy(out, p.conflictDescriptions)
	return out
}
