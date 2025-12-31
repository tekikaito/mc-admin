package services

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// UserCacheEntry represents one entry from usercache.json
type UserCacheEntry struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	ExpiresOn string `json:"expiresOn"`
}

// PlayerStats holds parsed stats for a player
type PlayerStats struct {
	UUID        string
	Name        string
	Stats       map[string]map[string]int64
	DataVersion int
}

// UserStatsService reads the usercache and per-player stats files
type UserStatsService struct {
	minecraftDataDir string
	// cached mapping from name->uuid and uuid->name
	nameToUUID map[string]string
	uuidToName map[string]string
}

// NewUserStatsService constructs the service and attempts to load the usercache.json
func NewUserStatsService(minecraftDataDir string) (*UserStatsService, error) {
	if minecraftDataDir == "" {
		return nil, fmt.Errorf("MINECRAFT_DATA_DIR is empty")
	}
	s := &UserStatsService{
		minecraftDataDir: minecraftDataDir,
		nameToUUID:       map[string]string{},
		uuidToName:       map[string]string{},
	}
	if err := s.loadUsercache(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *UserStatsService) loadUsercache() error {
	path := filepath.Join(s.minecraftDataDir, "usercache.json")
	buf, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var entries []UserCacheEntry
	if err := json.Unmarshal(buf, &entries); err != nil {
		return err
	}
	for _, e := range entries {
		s.nameToUUID[e.Name] = e.UUID
		s.uuidToName[e.UUID] = e.Name
	}
	return nil
}

// GetUUIDForName returns the uuid for the given username, ok=false if not found
func (s *UserStatsService) GetUUIDForName(name string) (string, bool) {
	u, ok := s.nameToUUID[name]
	return u, ok
}

// GetNameForUUID returns the username for the given uuid, ok=false if not found
func (s *UserStatsService) GetNameForUUID(uuid string) (string, bool) {
	n, ok := s.uuidToName[uuid]
	return n, ok
}

// GetStatsForUUID reads and parses world/stats/<uuid>.json
func (s *UserStatsService) GetStatsForUUID(uuid string) (*PlayerStats, error) {
	statsPath := filepath.Join(s.minecraftDataDir, "world", "stats", uuid+".json")
	// also try without the world prefix for setups that store stats under <datadir>/stats
	if _, err := os.Stat(statsPath); os.IsNotExist(err) {
		alt := filepath.Join(s.minecraftDataDir, "stats", uuid+".json")
		if _, err2 := os.Stat(alt); err2 == nil {
			statsPath = alt
		}
	}
	f, err := os.Open(statsPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// decode with UseNumber to preserve integer values
	var raw map[string]interface{}
	dec := json.NewDecoder(f)
	dec.UseNumber()
	if err := dec.Decode(&raw); err != nil && err != io.EOF {
		return nil, err
	}

	stats := map[string]map[string]int64{}
	if rawStats, ok := raw["stats"].(map[string]interface{}); ok {
		for category, v := range rawStats {
			inner := map[string]int64{}
			if m, ok := v.(map[string]interface{}); ok {
				for k, val := range m {
					switch num := val.(type) {
					case json.Number:
						if i64, err := num.Int64(); err == nil {
							inner[k] = i64
						}
					case float64:
						inner[k] = int64(num)
					case int:
						inner[k] = int64(num)
					case int64:
						inner[k] = num
					}
				}
			}
			stats[category] = inner
		}
	}

	dataVersion := 0
	if dv, ok := raw["DataVersion"]; ok {
		switch num := dv.(type) {
		case json.Number:
			if i64, err := num.Int64(); err == nil {
				dataVersion = int(i64)
			}
		case float64:
			dataVersion = int(num)
		case int:
			dataVersion = num
		case int64:
			dataVersion = int(num)
		}
	}

	name, _ := s.GetNameForUUID(uuid)

	return &PlayerStats{
		UUID:        uuid,
		Name:        name,
		Stats:       stats,
		DataVersion: dataVersion,
	}, nil
}

// GetStatsForName looks up the uuid and reads its stats
func (s *UserStatsService) GetStatsForName(name string) (*PlayerStats, error) {
	uuid, ok := s.GetUUIDForName(name)
	if !ok {
		return nil, fmt.Errorf("username not found: %s", name)
	}
	return s.GetStatsForUUID(uuid)
}

// GetAllPlayerStats returns stats for all users present in usercache.json (skips missing files)
func (s *UserStatsService) GetAllPlayerStats() ([]*PlayerStats, error) {
	var out []*PlayerStats
	for uuid := range s.uuidToName {
		ps, err := s.GetStatsForUUID(uuid)
		if err != nil {
			// skip missing/unreadable files
			continue
		}
		out = append(out, ps)
	}
	return out, nil
}
