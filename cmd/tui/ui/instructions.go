package ui

const usageGuide = `idp API Monitor — Quick Guide

1) Start the API (default :3000) and WebSocket hub (default :9090).
2) Open Auth+User → select Login.
3) Fill the login form fields and submit with Enter.
   • Access + refresh tokens are stored in memory automatically.
   • WebSocket auto-subscribes to deployment-log-stream-event.
4) Machine+Network → Create Machine opens an interactive form (name, CPU, memory, image: struckchure/alpine or struckchure/ubuntu).
   • Create Network, Repo Connection, and Deploy Repo use machine/repo dropdowns.
   • Deploy Repo auto-subscribes and streams deployment logs in the output panel.
5) Use other tabs for deployments, GitHub, etc.
   • Authenticated requests send Authorization: Bearer <token>.

Navigation
  shift+← / shift+→   switch section (tab)
  ↑ / ↓               select action
  tab / shift+tab     config fields (HTTP → WS) and form fields (↑/↓ in dropdowns)
  enter               open/submit interactive action form
  ctrl+s              force-submit current form
  o                   open latest link from API output (e.g. GitHub authorize link)
  ?                   toggle this guide in the output panel
  ctrl+l              clear output
  q / ctrl+c          quit

Config
  HTTP   API base URL (default http://localhost:3000)
  WS     WebSocket URL (default ws://localhost:9090/ws)

WebSocket events (subscribe manually on WebSocket tab)
  deployment-log-stream-event
  deployment-notification-event/<machineId>

Default admin (from .env): admin@idp.local / admin123
`

const defaultWSEvent = "deployment-log-stream-event"
