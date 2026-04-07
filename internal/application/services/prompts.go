package services

import _ "embed"

// commentPromptTemplate is loaded from comment_prompt.txt at build time.
// To customize: copy comment_prompt.txt.example → comment_prompt.txt (git-ignored).
//
//go:embed "comment_prompt.txt"
var commentPromptTemplate string
