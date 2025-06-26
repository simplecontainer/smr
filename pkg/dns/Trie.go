package dns

type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
}

type Trie struct {
	root *TrieNode
}

func NewTrie() *Trie {
	return &Trie{root: &TrieNode{children: make(map[rune]*TrieNode)}}
}

func (t *Trie) Insert(suffix string) {
	node := t.root
	// Insert suffix in reverse order for suffix matching
	runes := []rune(suffix)
	for i := len(runes) - 1; i >= 0; i-- {
		char := runes[i]
		if _, ok := node.children[char]; !ok {
			node.children[char] = &TrieNode{children: make(map[rune]*TrieNode)}
		}
		node = node.children[char]
	}
	node.isEnd = true
}

func (t *Trie) EndsWithSuffix(query string) (string, bool) {
	node := t.root
	runes := []rune(query)

	// Traverse from end of query backwards
	for i := len(runes) - 1; i >= 0; i-- {
		char := runes[i]
		if _, ok := node.children[char]; !ok {
			return "", false
		}
		node = node.children[char]
		if node.isEnd {
			// Return the actual suffix that was matched
			return string(runes[:i]), true
		}
	}
	return "", false
}
