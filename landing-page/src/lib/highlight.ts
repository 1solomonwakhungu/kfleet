export type CodeLanguage = 'yaml' | 'shell' | 'json' | 'text'

export type CodeTokenKind = 'key' | 'string' | 'flag' | 'variable' | 'punctuation' | 'plain'

export interface CodeToken {
  readonly kind: CodeTokenKind
  readonly text: string
}

// One alternation per language; the name of the group that matched is the
// token kind. This is a tint for a handful of known snippets, not a parser.
const PATTERNS: Record<Exclude<CodeLanguage, 'text'>, RegExp> = {
  yaml: /(?<string>"[^"\n]*")|(?<key>[A-Za-z][\w.-]*(?=:))|(?<punctuation>[[\]{},:]|^\s*-(?=\s))/gm,
  shell:
    /(?<string>"[^"\n]*")|(?<variable>\b[A-Z][A-Z0-9_]*(?==))|(?<flag>(?<=\s)--?[A-Za-z][\w-]*)|(?<punctuation>\\$)/gm,
  json: /(?<key>"[^"\n]*"(?=\s*:))|(?<string>"[^"\n]*")|(?<punctuation>[[\]{},:])/g,
}

export function tokenize(code: string, language: CodeLanguage): CodeToken[] {
  if (language === 'text') {
    return [{ kind: 'plain', text: code }]
  }

  const tokens: CodeToken[] = []
  let cursor = 0

  for (const match of code.matchAll(PATTERNS[language])) {
    if (match.index > cursor) {
      tokens.push({ kind: 'plain', text: code.slice(cursor, match.index) })
    }
    const matchedGroup = Object.entries(match.groups ?? {}).find(([, value]) => value !== undefined)
    tokens.push({ kind: (matchedGroup?.[0] ?? 'plain') as CodeTokenKind, text: match[0] })
    cursor = match.index + match[0].length
  }

  if (cursor < code.length) {
    tokens.push({ kind: 'plain', text: code.slice(cursor) })
  }

  return tokens
}
