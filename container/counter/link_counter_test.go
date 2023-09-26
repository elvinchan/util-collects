package counter

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/elvinchan/util-collects/as"
)

func TestLinkCounter(t *testing.T) {
	// prepare
	type Case struct {
		key     string
		hits    int64
		linkIds []string
	}
	var cases []Case
	for i := 1; i <= 10; i++ {
		cases = append(cases, Case{
			key: fmt.Sprintf("test-key-%d", i),
		})
	}
	lc := NewLinkCounter(8, 8)
	for i := 1; i <= 100; i++ {
		linkId := fmt.Sprintf("test-link-id-%d", i)
		for j := len(cases); j > 0; j-- {
			if i%10 > j {
				continue
			}
			// test-key-1 test-link-id-1 1 test-link-id-21 21 ...
			// test-key-2 test-link-id-1 2 test-link-id-2 4 test-link-id-21 42 test-link-id-22 44 ...
			// ...
			lc.Add(cases[j-1].key, int64(i*j), linkId)
			cases[j-1].hits += int64(i * j)
			cases[j-1].linkIds = append(cases[j-1].linkIds, linkId)
		}
	}

	kcs := lc.CountList()
	as.Equal(t, len(kcs), 8)
	for i, kc := range kcs {
		as.Equal(t, kc.Key, cases[len(cases)-1-i].key)
		as.Equal(t, kc.Count, cases[len(cases)-1-i].hits, fmt.Sprintf("hits not match for index: %d", i))
	}

	type Result struct {
		key     string
		hits    int64
		linkIds []string
	}
	var results []Result
	// check
	lc.Range(func(key string, hits int64, linkIds []string) bool {
		results = append(results, Result{
			key,
			hits,
			linkIds,
		})
		return true
	})
	as.Equal(t, len(results), 8)
	for _, result := range results {
		keyIndex, _ := strconv.Atoi(strings.TrimPrefix(result.key, "test-key-"))
		as.True(t, keyIndex > 2)
		as.Equal(t, result.hits, cases[keyIndex-1].hits, fmt.Sprintf("hits not match for index: %d", keyIndex))
		as.Equal(t, len(result.linkIds), len(cases[keyIndex-1].linkIds), fmt.Sprintf("linkIds length wrong for index: %d", keyIndex))
		for _, rid := range result.linkIds {
			exist := false
			for _, id := range cases[keyIndex-1].linkIds {
				if rid == id {
					exist = true
					break
				}
			}
			as.True(t, exist, fmt.Sprintf("linkId not exist for index: %d", keyIndex))
		}
	}
}
