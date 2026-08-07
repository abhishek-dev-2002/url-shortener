package urlshortener

import (
"context"
"fmt"
"sync"

"github.com/abhishekmaurya/url-shortner/repo/repointerfaces"
"github.com/abhishekmaurya/url-shortner/utils"
)

// =========================================================================
// Short Code Generator — DB Sequence + Block Allocation
// =========================================================================
//
// DESIGN:
//   Uses a PostgreSQL sequence as the single source of truth for ID generation.
//   Each app instance pre-allocates a block of 10,000 IDs at a time, then serves
//   codes from memory until the block is exhausted. Then it fetches the next block.
//
// HOW IT WORKS:
//   1. On first call (or when current block runs out), call DB:
//      SELECT setval('shortcode_id_seq', nextval(...) + 9999)
//      This atomically claims IDs [start, start+9999] for this instance.
//   2. Serve IDs from the in-memory counter (start → start+blockSize-1).
//   3. Base62-encode the counter → short, URL-safe, fixed-length code.
//   4. When block exhausted → allocate next block from DB.
//
// WHY IT WON'T COLLIDE:
//   - The DB sequence is atomic — two concurrent instances get different blocks
//   - Within a block, the mutex ensures sequential assignment
//   - No two instances ever share the same ID range
//
// SCALABILITY:
//   - DB is hit only once every 10,000 requests (amortized cost ≈ 0)
//   - 100 instances can run simultaneously — each gets non-overlapping blocks
//   - No coordination between instances beyond the shared sequence
//
// SHORT CODE LENGTH:
//   - Base62 encoding: 62^7 = 3.5 trillion possible codes
//   - At 1M URLs/day, that's ~9,500 years before we need 8 chars
//   - Codes start at 7 chars and stay there for a very long time
//
// TRADE-OFFS vs SNOWFLAKE:
//   ✅ Shorter codes (7 chars vs 10-11 for Snowflake)
//   ✅ No machine ID configuration needed — the DB sequence handles coordination
//   ✅ Simpler mental model — just an incrementing counter
//   ✅ Codes are compact and non-time-revealing
//   ⚠️  Requires a single shared DB (the sequence) — DB is the coordination point
//   ⚠️  If instance crashes mid-block, those unused IDs are "wasted" (acceptable)
//   ⚠️  Sequential codes are somewhat predictable (not a security concern for URLs)
//
// TRADE-OFFS vs PURE RANDOM:
//   ✅ Zero collisions by design (no retry loop, no existence check)
//   ✅ Much faster — no DB round-trip per code, only per block
//   ✅ Shorter codes guaranteed (sequential use of keyspace)
//   ⚠️  Codes are sequential (random would be unguessable)
//
// WHY BLOCK SIZE = 10,000:
//   - At 100 req/sec: one DB call every ~100 seconds
//   - At 1000 req/sec: one DB call every ~10 seconds
//   - Worst case waste on crash: 10,000 IDs = negligible in 3.5T keyspace
//   - Small enough to not starve other instances of the ID space
// =========================================================================

const (
// BlockSize is the number of IDs allocated from DB per batch.
BlockSize int64 = 10_000

// Base62 alphabet for encoding.
base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// MinCodeLength pads short codes to at least this many characters.
MinCodeLength = 7
)

// CodeGenerator allocates blocks of IDs from the database sequence
// and serves short codes from memory.
type CodeGenerator struct {
	mu      sync.Mutex
	repo    repointerfaces.URLRepository
	current int64 // current ID to serve
	ceiling int64 // upper bound of current block (exclusive)
}

// NewCodeGenerator creates a new block-allocating code generator.
func NewCodeGenerator(repo repointerfaces.URLRepository) *CodeGenerator {
	return &CodeGenerator{
		repo:    repo,
		current: 0,
		ceiling: 0, // forces allocation on first Generate()
	}
}

// Generate returns the next unique short code.
// Allocates a new block from DB when the current one is exhausted.
func (g *CodeGenerator) Generate(ctx context.Context) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// If current block is exhausted, fetch a new one
	if g.current >= g.ceiling {
		if err := g.allocateBlock(ctx); err != nil {
			return "", err
		}
	}

	id := g.current
	g.current++

	return encodeBase62(id), nil
}

// allocateBlock fetches the next block of IDs from the database.
func (g *CodeGenerator) allocateBlock(ctx context.Context) error {
	start, err := g.repo.AllocateIDBlock(ctx, BlockSize)
	if err != nil {
		utils.Error("failed to allocate ID block", "error", err)
		return fmt.Errorf("failed to allocate ID block: %w", err)
	}

	g.current = start
	g.ceiling = start + BlockSize

	utils.Info("allocated new ID block", "start", start, "end", g.ceiling-1)
	return nil
}

// encodeBase62 converts a positive integer to a Base62 string,
// padded to at least MinCodeLength characters.
func encodeBase62(num int64) string {
	if num == 0 {
		result := make([]byte, MinCodeLength)
		for i := range result {
			result[i] = base62Chars[0]
		}
		return string(result)
	}

	buf := make([]byte, 0, MinCodeLength)
	for num > 0 {
		buf = append(buf, base62Chars[num%62])
		num /= 62
	}

	// Pad to minimum length
	for len(buf) < MinCodeLength {
		buf = append(buf, base62Chars[0])
	}

	// Reverse (most significant first)
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}

	return string(buf)
}
