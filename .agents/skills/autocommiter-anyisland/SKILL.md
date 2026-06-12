# autocommiter-anyisland

Specialized skill for managing the integration between Autocommiter and the Anyisland ecosystem.

## Instructions

Use this skill when diagnosing integration issues with Anyisland or when managing the auto-registration process.

### Workflows

#### 1. Auto-Registration
- Autocommiter attempts to register with the local Anyisland daemon on every startup via `anyisland.Register()`.
- **Protocol**: UDP packet sent to `localhost:1995`.
- **Payload**: JSON containing `op` (REGISTER), `name`, `source`, `version`, and `type`.

#### 2. Management Status
- Check if the tool is currently "managed" by Anyisland Pulse using `autocommiter get-config`.
- **Heuristic**: It checks for a Unix socket at `~/.anyisland/anyisland.sock`.
- **Handshake**: Sends `{"op": "HANDSHAKE"}` and expects a `ManagedStatus` JSON response.

#### 3. Anyisland Pulse
- When managed, Autocommiter can receive heartbeats or instructions from the Pulse daemon, ensuring it stays updated and correctly configured within the island.
- **Updates**: Anyisland Pulse centralizes update orchestration, replacing the need for internal update logic in the binary.

### Key Components
- `internal/anyisland/anyisland.go`: Core registration and status logic.
- `cmd/autocommiter/main.go`: Calls `anyisland.Register()` on startup.
