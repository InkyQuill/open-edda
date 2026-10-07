package fileproject

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestConcurrentMetadataCreationHasOneWinner(t *testing.T) {
	root := t.TempDir()
	var wg sync.WaitGroup
	winners := make(chan string, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			id := fmt.Sprintf("project-%d", i)
			if _, err := InitMetadata(root, InitMetadataInput{ID: id, Title: id}); err == nil {
				winners <- id
			}
		}(i)
	}
	wg.Wait()
	close(winners)
	var ids []string
	for id := range winners {
		ids = append(ids, id)
	}
	if len(ids) != 1 {
		t.Fatalf("winners=%v", ids)
	}
	got, err := ReadMetadata(root)
	if err != nil || got.ID != ids[0] {
		t.Fatalf("metadata=%#v err=%v", got, err)
	}
}

func TestConcurrentConflictPreservationKeepsOneCompleteSet(t *testing.T) {
	root := t.TempDir()
	var wg sync.WaitGroup
	winners := make(chan string, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			body := fmt.Sprintf("variant-%d", i)
			_, err := PreserveConflict(root, PreserveConflictInput{FileID: "one", Path: "chapter.md", BaseMarkdown: body, LocalMarkdown: body, ServerMarkdown: body})
			if err == nil {
				winners <- body
			}
		}(i)
	}
	wg.Wait()
	close(winners)
	var bodies []string
	for body := range winners {
		bodies = append(bodies, body)
	}
	if len(bodies) != 1 {
		t.Fatalf("winners=%v", bodies)
	}
	for _, name := range []string{"base.md", "local.md", "server.md"} {
		body, err := os.ReadFile(filepath.Join(conflictDir(root, "one"), name))
		if err != nil || string(body) != bodies[0] {
			t.Fatalf("%s=%q err=%v", name, body, err)
		}
	}
}
