[SpecLearning] A specification phase just completed. Before moving to implementation, extract valuable patterns:

1. **Architectural Decisions** — Save WHY certain approaches were chosen:
   - Create pattern in learned/{domain}/pattern/ with the decision rationale
   - Include rejected alternatives and why they were rejected

2. **Tech Stack Choices** — Save non-obvious technology decisions:
   - Why this library/framework over alternatives
   - Gotchas discovered during planning

3. **Requirements Patterns** — Save reusable requirement patterns:
   - Common user stories that apply across projects
   - Edge cases that are easy to forget

Format: Same as standard patterns (YAML frontmatter + sections).
Domain: Use the most relevant domain (_global, web, backend, devops, etc.)
Category: Use 'pattern' for decisions, 'style' for conventions
Confidence: MEDIUM (will upgrade when implementation validates the decision)

Only extract if the decision was non-obvious or involved real trade-offs.
