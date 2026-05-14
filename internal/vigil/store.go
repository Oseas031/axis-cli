package vigil

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type Store struct {
	root string
}

func NewStore(root string) *Store {
	return &Store{root: root}
}

func (s *Store) dir() string {
	return filepath.Join(s.root, ".axis", "vigil")
}

func (s *Store) path() string {
	return filepath.Join(s.dir(), "items.json")
}

func (s *Store) Load() ([]*Item, error) {
	data, err := os.ReadFile(s.path())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []*Item{}, nil
		}
		return nil, err
	}
	var items []*Item
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Store) Save(items []*Item) error {
	if items == nil {
		items = []*Item{}
	}
	if err := os.MkdirAll(s.dir(), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path(), data, 0o644)
}

func (s *Store) Add(item *Item) error {
	if item == nil {
		return errors.New("item is nil")
	}
	items, err := s.Load()
	if err != nil {
		return err
	}
	items = append(items, item)
	return s.Save(items)
}

func (s *Store) Get(id string) (*Item, error) {
	if id == "" {
		return nil, errors.New("id is empty")
	}
	items, err := s.Load()
	if err != nil {
		return nil, err
	}
	for _, it := range items {
		if it.ID == id {
			return it, nil
		}
	}
	return nil, fmt.Errorf("item %s not found", id)
}

func (s *Store) Update(item *Item) error {
	if item == nil {
		return errors.New("item is nil")
	}
	items, err := s.Load()
	if err != nil {
		return err
	}
	for i, it := range items {
		if it.ID == item.ID {
			items[i] = item
			return s.Save(items)
		}
	}
	return fmt.Errorf("item %s not found", item.ID)
}

func (s *Store) Archive(items []*Item) error {
	if len(items) == 0 {
		return nil
	}
	archiveDir := filepath.Join(s.dir(), "archive")
	if err := os.MkdirAll(archiveDir, 0o755); err != nil {
		return err
	}
	grouped := map[string][]*Item{}
	for _, it := range items {
		key := time.Now().Format("2006-01")
		if it.CompletedAt != nil {
			key = it.CompletedAt.Format("2006-01")
		}
		grouped[key] = append(grouped[key], it)
	}
	for key, batch := range grouped {
		p := filepath.Join(archiveDir, key+".json")
		var existing []*Item
		data, err := os.ReadFile(p)
		if err == nil {
			_ = json.Unmarshal(data, &existing)
		}
		existing = append(existing, batch...)
		out, err := json.MarshalIndent(existing, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(p, out, 0o644); err != nil {
			return err
		}
	}
	return nil
}
