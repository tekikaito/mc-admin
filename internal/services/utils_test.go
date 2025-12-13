package services

import "fmt"

type fakeRconClient struct {
	responses map[string]struct {
		out string
		err error
	}
	received []string
}

func (f *fakeRconClient) ExecuteCommand(cmd string) (string, error) {
	f.received = append(f.received, cmd)
	if r, ok := f.responses[cmd]; ok {
		return r.out, r.err
	}
	return "", fmt.Errorf("unexpected command: %s", cmd)
}

type fakeMojangChecker struct {
	existsMap map[string]bool
	errMap    map[string]error
}

func (f *fakeMojangChecker) CheckMojangUsernameExists(username string) (bool, error) {
	if err, ok := f.errMap[username]; ok {
		return false, err
	}
	if exists, ok := f.existsMap[username]; ok {
		return exists, nil
	}
	return false, fmt.Errorf("unexpected username: %s", username)
}

type fakeFileClient struct {
	files map[string]string
}

func (f *fakeFileClient) ReadFile(path string) (string, error) {
	if content, ok := f.files[path]; ok {
		return content, nil
	}
	return "", fmt.Errorf("file not found: %s", path)
}
