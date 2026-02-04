package git

import (
	"fmt"
	"strings"
)

// ConfigGet reads a single git config value.
func ConfigGet(key string) (string, error) {
	return Run("config", "--local", "--get", key)
}

// ConfigSet writes a git config value.
func ConfigSet(key, value string) error {
	return RunSilent("config", "--local", key, value)
}

// ConfigUnset removes a git config key.
func ConfigUnset(key string) error {
	return RunSilent("config", "--local", "--unset", key)
}

// ConfigRemoveSection removes an entire config section.
func ConfigRemoveSection(section string) error {
	return RunSilent("config", "--local", "--remove-section", section)
}

// ConfigGetRegexp returns all config entries matching a pattern.
// Each result is a key-value pair.
func ConfigGetRegexp(pattern string) ([][2]string, error) {
	out, err := Run("config", "--local", "--get-regexp", pattern)
	if err != nil {
		// No matches is not an error for our purposes
		if strings.Contains(err.Error(), "exit status 1") || out == "" {
			return nil, nil
		}
		return nil, err
	}
	if out == "" {
		return nil, nil
	}

	var results [][2]string
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, " ", 2)
		if len(parts) == 2 {
			results = append(results, [2]string{parts[0], parts[1]})
		}
	}
	return results, nil
}

// GetTrunk reads the configured trunk branch.
func GetTrunk() (string, error) {
	trunk, err := ConfigGet("st.trunk")
	if err != nil {
		return "", fmt.Errorf("st not initialized. Run 'st init' first")
	}
	return trunk, nil
}

// SetTrunk writes the trunk branch to config.
func SetTrunk(branch string) error {
	return ConfigSet("st.trunk", branch)
}

// GetStackParent reads the parent of a stacked branch.
func GetStackParent(branch string) (string, error) {
	key := fmt.Sprintf("stack.%s.parent", branch)
	return ConfigGet(key)
}

// SetStackParent writes the parent of a stacked branch.
func SetStackParent(branch, parent string) error {
	key := fmt.Sprintf("stack.%s.parent", branch)
	return ConfigSet(key, parent)
}

// RemoveStackSection removes the config section for a branch.
func RemoveStackSection(branch string) error {
	section := fmt.Sprintf("stack.%s", branch)
	return ConfigRemoveSection(section)
}

// GetRestackState reads restack-in-progress state.
func GetRestackState() (string, error) {
	return ConfigGet("st.restack-remaining")
}

// SetRestackState saves restack-in-progress state.
func SetRestackState(remaining string) error {
	if err := ConfigSet("st.restack-in-progress", "true"); err != nil {
		return err
	}
	return ConfigSet("st.restack-remaining", remaining)
}

// ClearRestackState removes restack-in-progress state.
func ClearRestackState() error {
	_ = ConfigUnset("st.restack-in-progress")
	_ = ConfigUnset("st.restack-remaining")
	return nil
}

// IsRestackInProgress checks if a restack is in progress.
func IsRestackInProgress() bool {
	val, err := ConfigGet("st.restack-in-progress")
	return err == nil && val == "true"
}
