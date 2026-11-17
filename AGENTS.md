# AGENTS.md

## Setup commands

- Install deps: `pnpm install`
- Start dev server: `pnpm dev`
- Run tests: `pnpm test`

## Code style

- TypeScript strict mode
- Single quotes, no semicolons
- Use functional patterns where possible
- **Tailwind CSS**: Use canonical classes instead of arbitrary values whenever possible (e.g., `border-2` instead of `border-[2px]`, `rounded` instead of `rounded-[4px]`, `text-base` instead of `text-[16px]`).
- **Icons/Graphics**: Do not use emojis in frontend development. Use proper icon libraries (like `lucide-react`) instead.


# Task Management & TODO Guidelines
- When viewing, querying, adding, updating, or deleting project tasks or TODOs, ALWAYS use the TODO MCP tools (`todo_get_tasks`, `todo_add_tasks`, `todo_update_tasks`, `todo_delete_tasks`, `todo_clear_category`, `todo_move_category`) on server `todo-mcp` or `todo-extension`.
- NEVER edit the `.todo` file directly with file modification tools.
