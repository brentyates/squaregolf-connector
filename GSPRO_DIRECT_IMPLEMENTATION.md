# GSPro Direct Integration - Implementation Status

## Overview

This document describes the implementation of dual-mode GSPro integration, allowing the connector to work with either:
1. **GSPro Connect** (current/stable) - Connects to the GSPro Connect middleware
2. **GSPro Direct** (experimental) - Direct integration with GSPro, bypassing GSPro Connect

## Completed Implementation

### 1. Core Infrastructure ✅

**File: `internal/core/gspro/gspro_direct.go`**
- Created `DirectIntegration` struct with full skeleton implementation
- Implements all core methods: `Start()`, `Stop()`, `Connect()`, `Disconnect()`
- TCP connection management with keepalive
- Automatic reconnection with exponential backoff
- Message receive loop with JSON parsing (placeholder)
- State listener registration (placeholder)

**Placeholders for reverse engineering:**
- `sendInitialHandshake()` - Protocol handshake TBD
- `processMessage()` - Message types TBD
- `SendShotData()` - Shot data format TBD
- Port number (currently 921, likely different)

### 2. Configuration Support ✅

**File: `internal/config/config.go`**
- Added `GSProMode` field to `Settings` struct
- Default value: `"connect"`
- Added `SetGSProMode()` method for runtime configuration
- Persisted to `~/.squaregolf-connector/config.json`

### 3. Integration Layer ✅

**File: `internal/core/gspro/integration.go`**
- Updated `Integration` struct with:
  - `mode` field ("connect" or "direct")
  - `directImpl` field (holds `*DirectIntegration`)
- Added `SetMode()` method to switch between modes
- Added `GetMode()` method to query current mode
- Updated `Start()` to delegate to Direct mode when enabled
- Updated `Stop()` to delegate to Direct mode when enabled

### 4. Command-Line Interface ✅

**File: `main.go`**
- Added `--gspro-mode` flag (values: "connect" or "direct")
- Added `GSProMode` field to `AppConfig` struct
- Updated GSPro initialization in both headless and web modes
- Mode is applied before starting the integration

**Usage:**
```bash
# Use GSPro Connect (default)
./squaregolf-connector --enable-gspro

# Use GSPro Direct (experimental)
./squaregolf-connector --enable-gspro --gspro-mode=direct

# With custom IP/port
./squaregolf-connector --enable-gspro --gspro-mode=direct --gspro-ip=127.0.0.1 --gspro-port=9210
```

## Architecture

```
┌─────────────────────────────────────┐
│      SquareGolf Connector           │
│                                      │
│  ┌────────────────────────────────┐ │
│  │   Integration (Singleton)      │ │
│  │   - mode: "connect"|"direct"   │ │
│  └───────────┬────────────────────┘ │
│              │                       │
│       ┌──────┴──────┐               │
│       │             │               │
│  ┌────▼────┐  ┌────▼─────────┐     │
│  │ Connect │  │ Direct       │     │
│  │ (Port   │  │ (Port TBD)   │     │
│  │  921)   │  │              │     │
│  └────┬────┘  └────┬─────────┘     │
│       │            │                │
└───────┼────────────┼────────────────┘
        │            │
        ▼            ▼
  ┌─────────┐  ┌─────────┐
  │ GSPro   │  │ GSPro   │
  │ Connect │  │ (Direct)│
  └────┬────┘  └─────────┘
       │
       ▼
  ┌─────────┐
  │ GSPro   │
  └─────────┘
```

## Pending Implementation

### 5. Web API Endpoints ⏳

**File: `internal/web/server.go`** (Not yet implemented)

Need to add:
- `GET /api/gspro/mode` - Get current mode
- `POST /api/gspro/mode` - Set mode (requires restart)
- Update `GET /api/gspro/config` to include mode
- Update `POST /api/gspro/config` to accept mode

### 6. Frontend UI ⏳

**Files: `web/static/js/services/GSProService.js`, `web/static/index.html`** (Not yet implemented)

Need to add:
- Mode selection toggle/dropdown in GSPro settings
- Display current mode in status
- Warning when switching modes (requires reconnection)

### 7. Reverse Engineering ⏳

**See: `GSPRO_REVERSE_ENGINEERING.md`**

Must determine via reverse engineering:
1. **Port Number** - What port does GSPro listen on internally?
2. **Handshake Protocol** - Initial connection sequence
3. **Message Format** - JSON, binary, protobuf, or other?
4. **Message Types** - All message types GSPro expects/sends
5. **Shot Data Format** - Does it differ from GSPro Connect?

## Testing Plan

### Phase 1: Reverse Engineering (Current)
1. Locate GSPro Connect executable in Wine prefix
2. Decompile with dnSpy or ILSpy
3. Extract protocol details
4. Validate with network capture

### Phase 2: Protocol Implementation
1. Update `sendInitialHandshake()` in `gspro_direct.go`
2. Update `processMessage()` with actual message types
3. Update `SendShotData()` with correct format
4. Update port number constant

### Phase 3: Integration Testing
1. Test connection to GSPro without GSPro Connect
2. Test shot data transmission
3. Test player information updates
4. Test reconnection logic
5. Test error handling

### Phase 4: UI/API Completion
1. Implement API endpoints
2. Implement frontend mode selection
3. Add mode persistence
4. Update documentation

### Phase 5: User Testing
1. Test with real GSPro installation
2. Compare Direct vs Connect modes
3. Verify no regressions in Connect mode
4. Performance comparison

## Current State

✅ **Infrastructure Complete** - Can run in either mode
✅ **Configuration Complete** - Mode selection via CLI and config file
⏳ **Protocol Unknown** - Needs reverse engineering
⏳ **Web UI** - Not yet implemented

## Next Steps

1. **Immediate**: Reverse engineer GSPro Connect
   - Find executable in Wine prefix
   - Decompile with ILSpy/dnSpy
   - Document protocol

2. **After Reverse Engineering**: Implement protocol
   - Update placeholders in `gspro_direct.go`
   - Test connection
   - Verify shot data

3. **Polish**: Complete UI
   - Add API endpoints
   - Update frontend
   - Add documentation

## Benefits of Direct Mode

Once implemented, Direct mode will provide:
- ✅ No dependency on GSPro Connect bugs
- ✅ Simpler setup (one less process to manage)
- ✅ Potentially lower latency
- ✅ Direct control over protocol
- ✅ Easier debugging

## Backward Compatibility

The implementation maintains full backward compatibility:
- Default mode is "connect" (existing behavior)
- No changes required to use Connect mode
- Users can opt-in to Direct mode when ready
- Both modes use the same state management and UI

## Files Modified

1. `internal/core/gspro/gspro_direct.go` - **NEW** - Direct integration implementation
2. `internal/core/gspro/integration.go` - Mode selection and delegation
3. `internal/config/config.go` - GSProMode configuration
4. `main.go` - Command-line flag and mode initialization
5. `GSPRO_REVERSE_ENGINEERING.md` - **NEW** - Reverse engineering plan
6. `GSPRO_DIRECT_IMPLEMENTATION.md` - **NEW** - This document

## Known Limitations

- Direct mode protocol is not yet implemented (all TODOs)
- No web UI for mode selection yet
- Port number is placeholder (921, likely incorrect)
- Message formats are placeholders
- State listeners not fully wired up for Direct mode

## Questions to Answer via Reverse Engineering

1. What TCP port does GSPro use internally?
2. Is the protocol JSON, binary, or something else?
3. Does GSPro require authentication/registration?
4. What's the handshake sequence?
5. Are message formats identical to GSPro Connect API?
6. Does GSPro use keep-alive messages?
7. Are there version compatibility checks?
8. What error codes/messages exist?

---

**Status**: Infrastructure complete, awaiting reverse engineering to fill in protocol details.
