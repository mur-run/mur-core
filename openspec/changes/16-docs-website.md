# 16: Documentation Website

## Summary
Create documentation website structure for docs.murmur-ai.dev using MkDocs with Material theme.

## Motivation
- README is getting long; need structured, searchable documentation
- Need detailed command references, concepts, and integration guides
- Better onboarding experience for new users

## Design

### Structure
```
docs/
├── index.md              # Homepage
├── getting-started/
│   ├── installation.md
│   ├── quick-start.md
│   └── configuration.md
├── commands/
│   ├── run.md
│   ├── sync.md
│   ├── learn.md
│   ├── stats.md
│   └── team.md
├── concepts/
│   ├── patterns.md
│   ├── routing.md
│   └── cross-cli-sync.md
├── integrations/
│   ├── claude-code.md
│   ├── gemini-cli.md
│   ├── auggie.md
│   └── openclaw.md
└── pricing.md
```

### Technology
- **MkDocs** with Material theme
- **GitHub Pages** for hosting
- Custom domain: docs.murmur-ai.dev

### Features
- Full-text search
- Mobile responsive
- Dark/light mode
- Copy code buttons
- Navigation tabs

## Implementation

1. Install MkDocs: `pip install mkdocs-material`
2. Create `mkdocs.yml` configuration
3. Write documentation content
4. Setup GitHub Pages workflow
5. Configure DNS for custom domain

## Deployment
GitHub Actions workflow triggers on push to main, builds docs, deploys to GitHub Pages.

## Status
- [x] Change spec created
- [x] MkDocs configuration
- [x] Documentation content
- [x] GitHub Actions workflow
- [ ] DNS configuration (manual setup required)
